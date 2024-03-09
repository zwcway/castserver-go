#include "./samples.h"

int config_buffer_duration = 10;

int cs_samples_init(CS_Samples *s)
{
    s->format = s->real_fmt;

    s->raw_data = (uint8_t **)av_malloc(s->real_size + BUFFER_OFFSET);
    if (!s->data)
        return AVERROR(ENOMEM);

    s->data = (double **)s->raw_data;

    // 构造 ffmpeg 支持的二维数组
    uint8_t *buf = (uint8_t *)s->raw_data + BUFFER_OFFSET;
    int per_ch_size = cs_format_size(s->format, s->req_nb_samples);
    for (int ch = 0; ch < s->format.chs; ch++, buf += per_ch_size)
        s->raw_data[ch] = buf;

    // 构建临时指针
    buf = (uint8_t *)s->raw_data + BUFFER_OFFSET;
    for (int ch = 0; ch < s->format.chs; ch++, buf += per_ch_size)
        s->ptr[ch] = buf;

    return 0;
}

int cs_samples_copy_from(CS_Samples *s, const uint8_t *const src, int nb_samples, const CS_Format f)
{
    int size = cs_format_size(f, nb_samples);

    if (size > s->real_size)
        return 1;

    memcpy(s->raw_data + BUFFER_OFFSET, src, size);
    memcpy(s->ptr, s->raw_data, BUFFER_OFFSET);

    s->format = f;

    s->last_nb_samples = nb_samples;
    s->req_nb_samples = nb_samples;
}

void cs_samples_reset(CS_Samples *s)
{
    // 重置临时指针的初始位置
    memcpy(s->ptr, s->raw_data, BUFFER_OFFSET);
}

void cs_samples_copy_step(CS_Samples *s, void *dst, int ch, int index)
{
    uint8_t *sd;
    switch (s->format.bitw)
    {
    case 1:
        *((uint8_t *)dst) = ((uint8_t *)(s->raw_data[ch]))[index];
        break;
    case 2:
        *((uint16_t *)dst) = ((uint16_t *)(s->raw_data[ch]))[index];
        break;
    case 3:
        sd = ((uint8_t *)(s->raw_data[ch])) + index * 3;
        *((uint8_t *)dst) = *sd;
        *(((uint8_t *)dst) + 1) = *(sd + 1);
        *(((uint8_t *)dst) + 2) = *(sd + 2);
        break;
    case 4:
        *((uint32_t *)dst) = ((uint32_t *)(s->raw_data[ch]))[index];
        break;
    case 8:
        *((uint64_t *)dst) = ((uint64_t *)(s->raw_data[ch]))[index];
        break;
    }
}

int cs_samples_defaultNBSamples(CS_Format f)
{
    return SAMPLES_BY_DURATION(f.srate, config_buffer_duration);
}

/**
 * @brief 创建样本缓存
 *
 * @param duration
 * @param rate
 * @param bit
 * @param chs
 * @return CS_Samples*
 */
CS_Samples *cs_create_samples(CS_Format f)
{
    int samples = cs_samples_defaultNBSamples(f);
    int size = cs_format_size(f, samples);
    if (size <= 0)
        return NULL;

    CS_Samples *s = (CS_Samples *)av_mallocz(size);
    if (!s)
        return NULL;

    s->real_fmt = f;
    s->real_size = size;
    s->req_nb_samples = samples;
    s->last_nb_samples = samples;

    if (!cs_samples_init(s))
    {
        av_free(s);
        return NULL;
    }

    return s;
}

void cs_samples_destory(CS_Samples **s)
{
    if (!s || !*s)
        return;

    av_freep(&(*s)->raw_data);
    (*s)->data = NULL;

    av_freep(s);
}