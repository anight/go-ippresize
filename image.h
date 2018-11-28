
#ifndef IMAGE_H
#define IMAGE_H 1

#include <stddef.h>

struct image_s {
	unsigned w, h;
	unsigned channels;
	size_t rowstep;
};

typedef enum {
	IMAGE_INTERPOLATION_NN,
	IMAGE_INTERPOLATION_LINEAR,
	IMAGE_INTERPOLATION_CUBIC,
	IMAGE_INTERPOLATION_LANCZOS,
	IMAGE_INTERPOLATION_SUPER,
	IMAGE_INTERPOLATION_ANTIALIASING_LINEAR,
	IMAGE_INTERPOLATION_ANTIALIASING_CUBIC,
	IMAGE_INTERPOLATION_ANTIALIASING_LANCZOS,
} image_interpolation_t;

void image_init();
image_interpolation_t image_interpolation_by_name(const char *name);
int image_ipp_resize(const struct image_s *in, const unsigned char *in_data, struct image_s *out, unsigned char *out_data, image_interpolation_t interpolation);
int image_ipp_replicate_border_inplace(struct image_s *dst_im, unsigned char *dst_im_data, unsigned src_off_x, unsigned src_off_y, unsigned src_w, unsigned src_h);

#endif

