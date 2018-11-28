
#include <string.h>

#include <ipp.h>

#include "image.h"

#define channels_select_C134R(channels, name) (channels == 1 ? name##_C1R : (channels == 3 ? name##_C3R : name##_C4R))
#define channels_select_C134IR(channels, name) (channels == 1 ? name##_C1IR : (channels == 3 ? name##_C3IR : name##_C4IR))

static int image_ipp_inter(image_interpolation_t inter, IppiInterpolationType *interpolation, int *antialiasing)
{
	switch (inter) {
		case IMAGE_INTERPOLATION_NN:
			*antialiasing = 0;
			*interpolation = ippNearest;
			return 0;
		case IMAGE_INTERPOLATION_LINEAR:
			*antialiasing = 0;
			*interpolation = ippLinear;
			return 0;
		case IMAGE_INTERPOLATION_CUBIC:
			*antialiasing = 0;
			*interpolation = ippCubic;
			return 0;
		case IMAGE_INTERPOLATION_LANCZOS:
			*antialiasing = 0;
			*interpolation = ippLanczos;
			return 0;
		case IMAGE_INTERPOLATION_SUPER:
			*antialiasing = 0;
			*interpolation = ippSuper;
			return 0;
		case IMAGE_INTERPOLATION_ANTIALIASING_LINEAR:
			*antialiasing = 1;
			*interpolation = ippLinear;
			return 0;
		case IMAGE_INTERPOLATION_ANTIALIASING_CUBIC:
			*antialiasing = 1;
			*interpolation = ippCubic;
			return 0;
		case IMAGE_INTERPOLATION_ANTIALIASING_LANCZOS:
			*antialiasing = 1;
			*interpolation = ippLanczos;
			return 0;
		default:
			return -1;
	}
}


int image_ipp_resize(const struct image_s *in, const unsigned char *in_data, struct image_s *out, unsigned char *out_data, image_interpolation_t inter)
{
	if (in->channels != 1 && in->channels != 3 && in->channels != 4) {
		return -1;
	}

	if (in->channels != out->channels) {
		return -1;
	}

	if (!out_data) {
		return -1;
	}

	/* special parameters for Lanczos */
	const Ipp32u numLobes = 3;

	/* special parameters for Cubic */
	/* ippi.h: ippCubic = IPPI_INTER_CUBIC2P_CATMULLROM */
	/* ippidefs.h: IPPI_INTER_CUBIC2P_CATMULLROM, // two-parameter cubic filter (B=0, C=1/2) */
	const Ipp32f valueB = 0.;
	const Ipp32f valueC = 0.5;

	int antialiasing;
	IppiInterpolationType interpolation;

	if (0 > image_ipp_inter(inter, &interpolation, &antialiasing)) {
		return -1;
	}

	IppStatus ippSts;

	IppiSize srcSize = { in->w, in->h };
	IppiSize dstSize = { out->w, out->h };

	int iSpecSize;
	int iInitSize;

	ippSts = ippiResizeGetSize_8u(srcSize, dstSize, interpolation, antialiasing, &iSpecSize, &iInitSize);
	if (ippSts != ippStsNoErr) {
		return -1;
	}

	IppiResizeSpec_32f *pSpec = (IppiResizeSpec_32f *) ippsMalloc_8u(iSpecSize);
	if (pSpec == NULL) {
		return -1;
	}

	Ipp8u *pInitBuf = 0;

	if (iInitSize) {
		pInitBuf = ippsMalloc_8u(iInitSize);
		if (pInitBuf == NULL) {
			ippsFree(pSpec);
			return -1;
		}
	}

	switch (interpolation) {
		case ippNearest:
			ippSts = ippiResizeNearestInit_8u(srcSize, dstSize, pSpec);
			break;
		case ippLinear:
			if (antialiasing) {
				ippSts = ippiResizeAntialiasingLinearInit(srcSize, dstSize, pSpec, pInitBuf);
			} else {
				ippSts = ippiResizeLinearInit_8u(srcSize, dstSize, pSpec);
			}
			break;
		case ippCubic:
			if (antialiasing) {
				ippSts = ippiResizeAntialiasingCubicInit(srcSize, dstSize, valueB, valueC, pSpec, pInitBuf);
			} else {
				ippSts = ippiResizeCubicInit_8u(srcSize, dstSize, valueB, valueC, pSpec, pInitBuf);
			}
			break;
		case ippLanczos:
			if (antialiasing) {
				ippSts = ippiResizeAntialiasingLanczosInit(srcSize, dstSize, numLobes, pSpec, pInitBuf);
			} else {
				ippSts = ippiResizeLanczosInit_8u(srcSize, dstSize, numLobes, pSpec, pInitBuf);
			}
			break;
		case ippSuper:
			ippSts = ippiResizeSuperInit_8u(srcSize, dstSize, pSpec);
			break;
		default:
			return -1;
	}

	ippsFree(pInitBuf);

	if (ippSts != ippStsNoErr) {
		ippsFree(pSpec);
		return -1;
	}

	int bufSize = 0;
	ippSts = ippiResizeGetBufferSize_8u(pSpec, dstSize, out->channels, &bufSize);
	if (ippSts != ippStsNoErr) {
		ippsFree(pSpec);
		return -1;
	}

	Ipp8u* pBuffer;
	pBuffer = ippsMalloc_8u(bufSize);
	if (pBuffer == NULL) {
		ippsFree(pSpec);
		return -1;
	}

	IppiPoint dstOffset = {0, 0};

	if (antialiasing) {
		ippSts = channels_select_C134R(in->channels, ippiResizeAntialiasing_8u)
			(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
	} else {
		switch (interpolation) {
			case ippNearest:
				ippSts = channels_select_C134R(in->channels, ippiResizeNearest_8u)
					(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, pSpec, pBuffer);
				break;
			case ippLinear:
				ippSts = channels_select_C134R(in->channels, ippiResizeLinear_8u)
					(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
				break;
			case ippCubic:
				ippSts = channels_select_C134R(in->channels, ippiResizeCubic_8u)
					(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
				break;
			case ippLanczos:
				ippSts = channels_select_C134R(in->channels, ippiResizeLanczos_8u)
					(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
				break;
			case ippSuper:
				ippSts = channels_select_C134R(in->channels, ippiResizeSuper_8u)
					(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, pSpec, pBuffer);
				break;
			default:
				return -1;
		}
	}

	ippsFree(pSpec);
	ippiFree(pBuffer);

	if (ippSts != ippStsNoErr) {
		return -1;
	}

	return 0;
}

int image_ipp_replicate_border_inplace(struct image_s *dst_im, unsigned char *dst_im_data, unsigned src_off_x, unsigned src_off_y, unsigned src_w, unsigned src_h)
{
	if (dst_im->channels != 1 && dst_im->channels != 3 && dst_im->channels != 4) {
		return -1;
	}

	IppStatus ippSts;
	unsigned char *start = dst_im_data + dst_im->channels * src_off_x + dst_im->rowstep * src_off_y;

	IppiSize srcSize = { src_w, src_h };
	IppiSize dstSize = { dst_im->w, dst_im->h };

	ippSts = channels_select_C134IR(dst_im->channels, ippiCopyReplicateBorder_8u)
		(start, dst_im->rowstep, srcSize, dstSize, src_off_y, src_off_x);

	if (ippSts != ippStsNoErr) {
		return -1;
	}

	return 0;
}

