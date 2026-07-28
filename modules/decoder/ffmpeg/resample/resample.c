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
    AVChannelLayout in_chs_layout;
    /**
     * @brief 指向 (uint8_t**) 偏移 256
     *
     */
    uint8_t *in_buffer;
    int in_buf_size;
    int in_linesize;       /* 缓存上次的 line size，供复用分支回填 */
    enum AVSampleFormat in_fmt;

    int out_rate;
    AVChannelLayout out_chs_layout;
    /**
     * @brief 指向 (uint8_t**) 偏移 256
     *
     */
    uint8_t *out_buffer;
    int out_buf_size;
    int out_linesize;      /* 缓存上次的 line size */
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

static const int go_swr_init(GOResample **ctx, int in_rate, const AVChannelLayout *in_chs_layout, enum AVSampleFormat in_fmt, int out_rate, const AVChannelLayout *out_chs_layout, enum AVSampleFormat out_fmt)
{
    if (ctx == NULL || in_chs_layout == NULL || out_chs_layout == NULL)
        return AVERROR(EINVAL);

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
    r->in_chs_layout = *in_chs_layout;
    r->in_fmt = in_fmt;

    r->out_rate = out_rate;
    r->out_chs_layout = *out_chs_layout;
    r->out_fmt = out_fmt;

    int ret = swr_alloc_set_opts2(&r->swrctx,
                       out_chs_layout, out_fmt, out_rate,
                       in_chs_layout, in_fmt, in_rate,
                       0, NULL);
    if (ret < 0) {
        go_swr_free(ctx);
        return ret;
    }

    ret = swr_init(r->swrctx);
    if (ret < 0)
    {
        go_swr_free(ctx);
        return ret;
    }

    return ret;
}

static const int malloc_in_buffer(GOResample *ctx, int nb_samples)
{
    int ret, chs = ctx->in_chs_layout.nb_channels;
    int inBufferSize = av_samples_get_buffer_size(NULL, chs, nb_samples, ctx->in_fmt, 0);
    if (ctx->in_buf_size >= inBufferSize) {
        return 0;
    }

    if (ctx->in_buffer != NULL)
        av_freep(&ctx->in_buffer);

    ctx->in_buf_size = inBufferSize;

    int line_size = 0;
    ret = av_samples_alloc_array_and_samples(
        (uint8_t ***)&ctx->in_buffer,
        &line_size,
        chs,
        nb_samples,
        ctx->in_fmt,
        0
    );

    if (ret < 0)
    {
        av_freep(&ctx->in_buffer);
        ctx->in_buf_size = 0;
        ctx->in_buffer = NULL;
        ctx->in_linesize = 0;
        return ret;
    }

    ctx->in_linesize = line_size;
    return ret;
}

static int expected_out_nb_samples(GOResample *ctx, int nb_samples)
{
    int64_t delay = swr_get_delay(ctx->swrctx, ctx->in_rate);
    return av_rescale_rnd(delay + nb_samples, ctx->out_rate, ctx->in_rate, AV_ROUND_UP);
}

static int go_convert(GOResample *ctx, int nb_samples)
{
    int ret, chs = ctx->out_chs_layout.nb_channels, out_linesize;
    int out_nb_samples = expected_out_nb_samples(ctx, nb_samples);

    // 缓冲区动态扩容
    int outBufferSize = av_samples_get_buffer_size(NULL, chs, out_nb_samples, ctx->out_fmt, 0);
    if (outBufferSize < 0)
        return outBufferSize;

    if (ctx->out_buf_size < outBufferSize)
    {
        if (ctx->out_buffer != NULL)
            av_freep(&ctx->out_buffer);

        ctx->out_buf_size = outBufferSize;

        ret = av_samples_alloc_array_and_samples(
            (uint8_t ***)&ctx->out_buffer,
            &out_linesize,
            chs,
            out_nb_samples,
            ctx->out_fmt,
            0
        );

        if (ret < 0)
        {
            av_freep(&ctx->out_buffer);
            ctx->out_buf_size = 0;
            ctx->out_buffer = NULL;
            ctx->out_linesize = 0;
            return ret;
        }

        ctx->out_linesize = out_linesize;
    }

    const uint8_t ** in_buf = (const uint8_t **)ctx->in_buffer;
    uint8_t ** out_buf = (uint8_t **)ctx->out_buffer;

    ret = swr_convert(ctx->swrctx, out_buf, out_nb_samples, in_buf, nb_samples);
    if (ret < 0)
        return ret;

    return ret;
}

/**
 * 获取指定声道的指针与 line size。
 * 用于 Go 端访问 planar 缓冲区：每个声道独立分配，长度等于 line_size 字节。
 */
static uint8_t *go_get_plane(uint8_t *planes, int ch)
{
    /* planes 是 av_samples_alloc_array_and_samples 写入的 uint8_t*[chs] 指针数组 */
    return ((uint8_t **)planes)[ch];
}

static int go_get_in_linesize(GOResample *ctx)
{
    return ctx->in_linesize;
}

static int go_get_out_linesize(GOResample *ctx)
{
    return ctx->out_linesize;
}
