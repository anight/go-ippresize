
#include <string.h>

#include <ipp.h>

#include "image.h"


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
		if (in->channels == 4) {
			ippSts = ippiResizeAntialiasing_8u_C4R(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
		} else {
			ippSts = (in->channels == 3 ? ippiResizeAntialiasing_8u_C3R : ippiResizeAntialiasing_8u_C1R)
				(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
		}
	} else {
		switch (interpolation) {
			case ippNearest:
				if (in->channels == 4) {
					ippSts = ippiResizeNearest_8u_C4R(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, pSpec, pBuffer);
				} else {
					ippSts = (in->channels == 3 ? ippiResizeNearest_8u_C3R : ippiResizeNearest_8u_C1R)
						(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, pSpec, pBuffer);
				}
				break;
			case ippLinear:
				if (in->channels == 4) {
					ippSts = ippiResizeLinear_8u_C4R(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
				} else {
					ippSts = (in->channels == 3 ? ippiResizeLinear_8u_C3R : ippiResizeLinear_8u_C1R)
						(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
				}
				break;
			case ippCubic:
				if (in->channels == 4) {
					ippSts = ippiResizeCubic_8u_C4R(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
				} else {
					ippSts = (in->channels == 3 ? ippiResizeCubic_8u_C3R : ippiResizeCubic_8u_C1R)
						(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
				}
				break;
			case ippLanczos:
				if (in->channels == 4) {
					ippSts = ippiResizeLanczos_8u_C4R(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
				} else {
					ippSts = (in->channels == 3 ? ippiResizeLanczos_8u_C3R : ippiResizeLanczos_8u_C1R)
						(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
				}
				break;
			case ippSuper:
				if (in->channels == 4) {
					ippSts = ippiResizeSuper_8u_C4R(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, pSpec, pBuffer);
				} else {
					ippSts = (in->channels == 3 ? ippiResizeSuper_8u_C3R : ippiResizeSuper_8u_C1R)
						(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, pSpec, pBuffer);
				}
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

