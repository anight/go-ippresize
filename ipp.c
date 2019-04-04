
#include <stdio.h>
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


const char *image_strerror(int code)
{
	switch (code) {
		case IMAGE_ERR_MEMORY_ALLOCATION_FAILED:
			return "Memory allocation failed";
		case IMAGE_ERR_INVALID_NUMBER_CHANNELS:
			return "Invalid number of image channels";
		case IMAGE_ERR_OUT_IMAGE_UNALLOCATED:
			return "Output image is unallocated";
		case IMAGE_ERR_INVALID_INTERPOLATION:
			return "Invalid interpolation";
		default:
			return ippGetStatusString(code);
	}
}


#define error_code_ipp(fmt...) error_code(ippSts, fmt)
#define error_code(code, fmt...) ({ \
	ippSts = code; \
	char temp[err_size]; \
	snprintf(temp, err_size, fmt); temp[err_size - 1] = '\0'; \
	snprintf(err, err_size, "%s: %s (%d)", temp, image_strerror(ippSts), ippSts); \
	(ippSts); \
})


int image_ipp_resize(const struct image_s *in, const unsigned char *in_data, struct image_s *out, unsigned char *out_data, image_interpolation_t inter, char *err, size_t err_size)
{
	IppStatus ippSts;

	if (in->channels != 1 && in->channels != 3 && in->channels != 4) {
		return error_code(IMAGE_ERR_INVALID_NUMBER_CHANNELS, "in->channels=%u", in->channels);
	}

	if (in->channels != out->channels) {
		return error_code(IMAGE_ERR_INVALID_NUMBER_CHANNELS, "in->channels=%u, out->channels=%u", in->channels, out->channels);
	}

	if (out_data == NULL) {
		return error_code(IMAGE_ERR_OUT_IMAGE_UNALLOCATED, "out_data == NULL");
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
		return error_code(IMAGE_ERR_INVALID_INTERPOLATION, "inter=%d", inter);
	}

	IppiSize srcSize = { in->w, in->h };
	IppiSize dstSize = { out->w, out->h };

	int iSpecSize;
	int iInitSize;

	ippSts = ippiResizeGetSize_8u(srcSize, dstSize, interpolation, antialiasing, &iSpecSize, &iInitSize);
	if (ippSts != ippStsNoErr) {
		return error_code_ipp("ippiResizeGetSize_8u() failed, srcSize={width: %d, height: %d}, dstSize={width: %d, height: %d}, inter=%d, antialiasing=%d",
			srcSize.width, srcSize.height, dstSize.width, dstSize.height, inter, antialiasing);
	}

	IppiResizeSpec_32f *pSpec = (IppiResizeSpec_32f *) ippsMalloc_8u(iSpecSize);
	if (pSpec == NULL) {
		return error_code(IMAGE_ERR_MEMORY_ALLOCATION_FAILED, "pSpec == NULL");
	}

	Ipp8u *pInitBuf = NULL;

	if (iInitSize) {
		pInitBuf = ippsMalloc_8u(iInitSize);
		if (pInitBuf == NULL) {
			ippsFree(pSpec);
			return error_code(IMAGE_ERR_MEMORY_ALLOCATION_FAILED, "pInitBuf == NULL");
		}
	}

	const char *init_function_name = NULL;

	switch (interpolation) {
		case ippNearest:
			ippSts = ippiResizeNearestInit_8u(srcSize, dstSize, pSpec);
			init_function_name = "ippiResizeNearestInit_8u";
			break;
		case ippLinear:
			if (antialiasing) {
				ippSts = ippiResizeAntialiasingLinearInit(srcSize, dstSize, pSpec, pInitBuf);
				init_function_name = "ippiResizeAntialiasingLinearInit";
			} else {
				ippSts = ippiResizeLinearInit_8u(srcSize, dstSize, pSpec);
				init_function_name = "ippiResizeLinearInit_8u";
			}
			break;
		case ippCubic:
			if (antialiasing) {
				ippSts = ippiResizeAntialiasingCubicInit(srcSize, dstSize, valueB, valueC, pSpec, pInitBuf);
				init_function_name = "ippiResizeAntialiasingCubicInit";
			} else {
				ippSts = ippiResizeCubicInit_8u(srcSize, dstSize, valueB, valueC, pSpec, pInitBuf);
				init_function_name = "ippiResizeCubicInit_8u";
			}
			break;
		case ippLanczos:
			if (antialiasing) {
				ippSts = ippiResizeAntialiasingLanczosInit(srcSize, dstSize, numLobes, pSpec, pInitBuf);
				init_function_name = "ippiResizeAntialiasingLanczosInit";
			} else {
				ippSts = ippiResizeLanczosInit_8u(srcSize, dstSize, numLobes, pSpec, pInitBuf);
				init_function_name = "ippiResizeLanczosInit_8u";
			}
			break;
		case ippSuper:
			ippSts = ippiResizeSuperInit_8u(srcSize, dstSize, pSpec);
			init_function_name = "ippiResizeSuperInit_8u";
			break;
		default:
			return error_code(IMAGE_ERR_INVALID_INTERPOLATION, "interpolation=%d", interpolation);
	}

	ippsFree(pInitBuf);

	if (ippSts != ippStsNoErr) {
		ippsFree(pSpec);
		return error_code_ipp("%s() failed, srcSize={width: %d, height: %d}, dstSize={width: %d, height: %d}, channels=%u",
			init_function_name, srcSize.width, srcSize.height, dstSize.width, dstSize.height, in->channels);
	}

	int bufSize = 0;
	ippSts = ippiResizeGetBufferSize_8u(pSpec, dstSize, out->channels, &bufSize);
	if (ippSts != ippStsNoErr) {
		ippsFree(pSpec);
		return error_code_ipp("ippiResizeGetBufferSize_8u() failed, dstSize={width: %d, height: %d}, channels=%u",
			dstSize.width, dstSize.height, out->channels);
	}

	Ipp8u* pBuffer;
	pBuffer = ippsMalloc_8u(bufSize);
	if (pBuffer == NULL) {
		ippsFree(pSpec);
		return error_code(IMAGE_ERR_MEMORY_ALLOCATION_FAILED, "pBuffer == NULL");
	}

	IppiPoint dstOffset = {0, 0};

	const char *resize_function_name = NULL;

	if (antialiasing) {
		ippSts = channels_select_C134R(in->channels, ippiResizeAntialiasing_8u)
			(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
		resize_function_name = "ippiResizeAntialiasing_8u";
	} else {
		switch (interpolation) {
			case ippNearest:
				ippSts = channels_select_C134R(in->channels, ippiResizeNearest_8u)
					(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, pSpec, pBuffer);
				resize_function_name = "ippiResizeNearest_8u";
				break;
			case ippLinear:
				ippSts = channels_select_C134R(in->channels, ippiResizeLinear_8u)
					(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
				resize_function_name = "ippiResizeLinear_8u";
				break;
			case ippCubic:
				ippSts = channels_select_C134R(in->channels, ippiResizeCubic_8u)
					(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
				resize_function_name = "ippiResizeCubic_8u";
				break;
			case ippLanczos:
				ippSts = channels_select_C134R(in->channels, ippiResizeLanczos_8u)
					(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, ippBorderRepl, 0, pSpec, pBuffer);
				resize_function_name = "ippiResizeLanczos_8u";
				break;
			case ippSuper:
				ippSts = channels_select_C134R(in->channels, ippiResizeSuper_8u)
					(in_data, in->rowstep, out_data, out->rowstep, dstOffset, dstSize, pSpec, pBuffer);
				resize_function_name = "ippiResizeSuper_8u";
				break;
			default:
				return IMAGE_ERR_INVALID_INTERPOLATION;
		}
	}

	ippsFree(pSpec);
	ippiFree(pBuffer);

	if (ippSts != ippStsNoErr) {
		return error_code_ipp("%s() failed", resize_function_name);
	}

	return ippStsNoErr;
}

int image_ipp_replicate_border_inplace(struct image_s *dst_im, unsigned char *dst_im_data, unsigned src_off_x, unsigned src_off_y, unsigned src_w, unsigned src_h, char *err, size_t err_size)
{
	IppStatus ippSts;

	if (dst_im->channels != 1 && dst_im->channels != 3 && dst_im->channels != 4) {
		return error_code(IMAGE_ERR_INVALID_NUMBER_CHANNELS, "dst_im->channels=%u", dst_im->channels);
	}

	unsigned char *start = dst_im_data + dst_im->channels * src_off_x + dst_im->rowstep * src_off_y;

	IppiSize srcSize = { src_w, src_h };
	IppiSize dstSize = { dst_im->w, dst_im->h };

	ippSts = channels_select_C134IR(dst_im->channels, ippiCopyReplicateBorder_8u)
		(start, dst_im->rowstep, srcSize, dstSize, src_off_y, src_off_x);

	if (ippSts != ippStsNoErr) {
		return error_code_ipp("ippiCopyReplicateBorder_8u() failed");
	}

	return ippStsNoErr;
}

