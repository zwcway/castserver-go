#ifndef CS_RESAMPLE
#define CS_RESAMPLE

#include <libswresample/swresample.h>
#include "samples.h"

typedef struct CS_Resample
{
    SwrContext *swrctx;

    int power;

    CS_Format in_format;
    enum AVSampleFormat in_fmt;
    int in_buf_size;

    CS_Format out_format;
    enum AVSampleFormat out_fmt;
    uint8_t *out_buffer;
    int out_buf_size;

} CS_Resample;

CS_Resample *cs_create_resample();
int cs_resample_convert(CS_Resample *ctx, CS_Samples *s);
int cs_resample_setFormat(CS_Resample *r, CS_Format in_format, CS_Format out_format);

#endif // CS_RESAMPLE