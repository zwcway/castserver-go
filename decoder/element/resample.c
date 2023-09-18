#include "resample.h"

void cs_resample_free(CS_Resample **ctx)
{
    if (ctx == NULL || *ctx == NULL)
        return;
    CS_Resample *c = *ctx;

    if (c->swrctx != NULL)
    {
        swr_close(c->swrctx);
        swr_free(&c->swrctx);
    }

    if (c->out_buffer != NULL)
        av_freep(&c->out_buffer);

    *ctx = NULL;
}

CS_Resample *cs_create_resample()
{
    CS_Resample *r = (CS_Resample *)av_mallocz(sizeof(CS_Resample));
    if (r == NULL)
    {
        return NULL;
    }

    if (r->swrctx == NULL)
    {
        // 初始化转码器
        r->swrctx = swr_alloc();
        if (r->swrctx == NULL)
        {
            cs_resample_free(&r);
            return NULL;
        }
    }

    return r;
}

/**
 * @brief 创建缓存
 * 
 * @param buf 
 * @param f 
 * @param nb_samples 
 * @return int 缓存大小
 */
static const int _cs_resample_malloc_buffer(uint8_t **buf, CS_Format f, int nb_samples)
{
    enum AVSampleFormat fmt = cs_format_bits_to_fmt(f.bit);
    int size = av_samples_get_buffer_size(NULL, f.chs, nb_samples, fmt, 0);

    if (!buf)
        return 0;
    if (*buf) 
        av_freep(buf);

    (*buf) = (uint8_t *)av_malloc(size + BUFFER_OFFSET);
    if (!*buf)
        return 0;

    int ret = av_samples_fill_arrays((uint8_t **)(*buf), NULL, (*buf) + BUFFER_OFFSET, f.chs, nb_samples, fmt, 0);
    if (ret < 0)
    {
        av_freep(buf);
        return 0;
    }

    return size;
}

/**
 * @brief 根据输入样本数量，计算输出样本数量
 * 
 * @param ctx 
 * @param in_nb_samples 
 * @return int 
 */
static int _cs_resample_output_nb_samples(CS_Resample *ctx, int in_nb_samples)
{
    int64_t delay = swr_get_delay(ctx->swrctx, ctx->in_format.rate);
    return av_rescale_rnd(delay + in_nb_samples, ctx->out_format.rate, ctx->in_format.rate, AV_ROUND_UP);
}

int cs_resample_setFormat(CS_Resample *r, CS_Format in_format, CS_Format out_format)
{
    if (r == NULL || r->swrctx == NULL)
        return AVERROR(ENOMEM);

    int in_nb_samples = cs_samples_default_samples(in_format);
    int out_nb_samples = _cs_resample_output_nb_samples(r, in_nb_samples);

    r->in_format = in_format;
    r->in_fmt = cs_format_bits_to_fmt(in_format.bit);
    r->in_buf_size = av_samples_get_buffer_size(NULL, in_format.chs, in_nb_samples, r->in_fmt, 0);

    r->out_format = out_format;
    r->out_fmt = cs_format_bits_to_fmt(out_format.bit);
    r->out_buf_size = _cs_resample_malloc_buffer(&r->out_buffer, out_format, out_nb_samples);

    if (!r->in_buf_size || !r->out_buf_size)
        return AVERROR(ENOMEM);

    swr_alloc_set_opts(r->swrctx,
                       r->out_format.layout, r->out_fmt, r->out_format.rate,
                       r->in_format.layout, r->in_fmt, r->in_format.rate,
                       0, NULL);

    int ret = swr_init(r->swrctx);
    return ret;
}

/**
 * @brief 开始转码
 * 
 * @param ctx 
 * @param nb_samples 
 * @return int 转码成功的样本数量
 */
int cs_resample_convert(CS_Resample *ctx, CS_Samples *s)
{
    uint8_t *buf;
    int ret, chs;
    int dst_nb_samples = _cs_resample_output_nb_samples(ctx, s->req_nb_samples);

    if (ctx->out_buffer == NULL)
        return AVERROR(EPERM);

    ret = swr_convert(ctx->swrctx,
                      (uint8_t **)ctx->out_buffer, dst_nb_samples,
                      (const uint8_t **)s->raw_data, s->req_nb_samples);
    if (ret < 0)
        return ret;

    cs_samples_copy_from(s, ctx->out_buffer + BUFFER_OFFSET, dst_nb_samples, ctx->out_format);

    return ret;
}
