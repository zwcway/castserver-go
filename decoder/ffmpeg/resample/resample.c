#include <libswresample/swresample.h>

/**
 * @brief 预留 32 个声道
 * 
 */
#define BUFFER_OFFSET 256
typedef struct GOResample
{
    SwrContext *swrctx;

    int in_rate;
    int64_t in_chs_layout;
    /**
     * @brief 指向 (uint8_t**) 偏移 256
     * 
     */
    uint8_t *in_buffer; 
    int in_buf_size;
    enum AVSampleFormat in_fmt;

    int out_rate;
    int64_t out_chs_layout;
    /**
     * @brief 指向 (uint8_t**) 偏移 256
     * 
     */
    uint8_t *out_buffer;
    int out_buf_size;
    enum AVSampleFormat out_fmt;

} GOResample;

static void go_swr_free(GOResample **ctx)
{
    if (ctx == NULL || *ctx == NULL)
        return;
    GOResample *c = *ctx;

    if (c->swrctx != NULL)
    {
        swr_close(c->swrctx);
        swr_free(&c->swrctx);
    }
    if (c->in_buffer != NULL)
        av_freep(&c->in_buffer);

    if (c->out_buffer != NULL)
        av_freep(&c->out_buffer);

    *ctx = NULL;
}

static const int go_swr_init(GOResample **ctx, int in_rate, int64_t in_chs_layout, enum AVSampleFormat in_bits, int out_rate, int64_t out_chs_layout, enum AVSampleFormat out_bits)
{
    if (ctx == NULL)
        return AVERROR(ENOMEM);

    GOResample *r = *ctx;

    if (r == NULL)
    {
        r = *ctx = (GOResample *)av_mallocz(sizeof(GOResample));
        if (r == NULL)
        {
            return AVERROR(ENOMEM);
        }
    }

    r->in_rate = in_rate;
    r->in_chs_layout = in_chs_layout;
    r->in_fmt = in_bits;

    r->out_rate = out_rate;
    r->out_chs_layout = out_chs_layout;
    r->out_fmt = out_bits;

    if (r->swrctx == NULL)
    {
        // 初始化转码器
        r->swrctx = swr_alloc();
        if (r->swrctx == NULL)
        {
            return -1;
        }
    }

    swr_alloc_set_opts(r->swrctx,
                       out_chs_layout, out_bits, out_rate,
                       in_chs_layout, in_bits, in_rate,
                       0, NULL);

    int ret = swr_init(r->swrctx);
    if (ret < 0)
    {
        go_swr_free(ctx);
        return ret;
    }

    return ret;
}

static const int malloc_in_buffer(GOResample *ctx, int nb_samples)
{
    uint8_t *buf;
    int ret, chs = av_get_channel_layout_nb_channels(ctx->in_chs_layout);
    int inBufferSize = av_samples_get_buffer_size(NULL, chs, nb_samples, ctx->in_fmt, 0);
    if (ctx->in_buf_size >= inBufferSize)
        return 0;

    if (ctx->in_buffer != NULL)
        av_freep(&ctx->in_buffer);

    ctx->in_buf_size = inBufferSize;
    ctx->in_buffer = (uint8_t *)av_malloc(inBufferSize + BUFFER_OFFSET);
    buf = ctx->in_buffer + BUFFER_OFFSET;

    ret = av_samples_fill_arrays((uint8_t **)ctx->in_buffer, NULL, buf, chs, nb_samples, ctx->in_fmt, 0);
    if (ret < 0)
    {
        ctx->in_buf_size = 0;
        av_freep(&ctx->in_buffer);
        return ret;
    }

    return ret;
}

static int out_nb_samples(GOResample *ctx, int nb_samples)
{
    int64_t delay = swr_get_delay(ctx->swrctx, ctx->in_rate);
    return av_rescale_rnd(delay + nb_samples, ctx->out_rate, ctx->in_rate, AV_ROUND_UP);
}

static int go_convert(GOResample *ctx, int nb_samples)
{
    uint8_t *buf;
    int ret, chs;
    int dst_nb_samples = out_nb_samples(ctx, nb_samples);
    // int dst_nb_samples = nb_samples;

    int outBufferSize = av_samples_get_buffer_size(NULL, ctx->out_chs_layout, dst_nb_samples, ctx->out_fmt, 0);
    if (outBufferSize < 0)
        return outBufferSize;

    if (ctx->in_buffer == NULL)
        return AVERROR(EPERM);

    if (ctx->out_buf_size < outBufferSize)
    {
        if (ctx->out_buffer != NULL)
            av_freep(&ctx->out_buffer);

        ctx->out_buf_size = outBufferSize;

        ctx->out_buffer = (uint8_t *)av_malloc(outBufferSize + BUFFER_OFFSET); 
        if (ctx->out_buffer == NULL)
            return AVERROR(ENOMEM);

        buf = ctx->out_buffer + BUFFER_OFFSET;

        chs = av_get_channel_layout_nb_channels(ctx->out_chs_layout);

        ret = av_samples_fill_arrays((uint8_t **)ctx->out_buffer, NULL, buf, chs, dst_nb_samples, ctx->out_fmt, 0);
        if (ret < 0)
        {
            ctx->out_buf_size = 0;
            av_freep(&ctx->out_buffer);
            return ret;
        }
        av_samples_set_silence((uint8_t **)ctx->out_buffer, 0, dst_nb_samples, chs, ctx->out_fmt);
    }

    ret = swr_convert(ctx->swrctx,
                      (uint8_t **)ctx->out_buffer, dst_nb_samples,
                      (const uint8_t **)ctx->in_buffer, nb_samples);
    if (ret < 0)
        return ret;

    return ret;
}
