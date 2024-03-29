#include "ele_mixer.h"

/**
 * @brief 混合两个相同采样率、相同位宽的样本，即混合声道
 *
 * @param m
 * @param dst
 * @param src
 * @return int
 */
int ele_mixer_mixin_channel(ELE_Mixer *m, CS_Samples *dst, CS_Samples *src)
{
    double *sd, *dd;
    int ch, i;

    for (ch = 0; ch < src->format.chs && ch < dst->format.chs; ch++)
    {
        sd = src->data[ch];
        dd = dst->data[ch];
        for (i = 0; i < src->req_nb_samples && i < dst->req_nb_samples; i++, sd++, dd++)
        {
            *dd += *sd;
        }
    }

    return 0;
}

int ele_mixer_add(ELE_Mixer *m, CS_Sourcer *src)
{
    if (!m || !src)
        return 1;

    _CS_Mixer_Sourcer *s = (_CS_Mixer_Sourcer *)malloc(sizeof(_CS_Mixer_Sourcer));
    s->cs = src;
    s->rs = cs_create_resample();
    m->sources[m->size++] = s;

    return 0;
}

/**
 * @brief 混音器主处理入口
 *
 * @param mp
 * @param s
 * @return int
 */
int ele_mixer_stream(void *mp, CS_Samples *s)
{
    if (!mp || !s)
        return AVERROR(ENOMEM);

    ELE_Mixer *m = (ELE_Mixer *)mp;

    s->format = m->format;

    for (_CS_Mixer_Sourcer *e, **ep = m->sources; ep && (e = *ep) && e->cs->ele; ep++)
    {
        cs_samples_slience(m->buf);
        e->cs->f_stream(e->cs->ele, m->buf);

        cs_resample_convert(e->rs, m->buf);

        ele_mixer_mixin_channel(m, s, m->buf);
    }

    return 0;
}

/**
 * @brief 更改混音器的输出格式
 *
 * @param m
 * @param f
 * @return int
 */
int ele_mixer_setOutputFormat(ELE_Mixer *m, CS_Format f)
{
    CS_Format sf, mf = {0};
    int err;

    if (m->buf)
    {
        cs_samples_destory(&m->buf);
    }

    cs_format_merge(&mf, f);
    m->format = f;

    // 重新设置每个转码器的格式
    for (_CS_Mixer_Sourcer *e, **ep = m->sources; ep && (e = *ep) && e->cs->ele; ep++)
    {
        // 获取输入格式
        e->cs->f_format(e->cs->ele, &sf);

        if (err = cs_resample_setFormat(e->rs, sf, f))
        {
            return err;
        }

        cs_format_merge(&mf, sf);
    }

    m->buf = cs_create_samples(mf);

    return 0;
}

CS_Streamer *ele_mixer_streamer(ELE_Mixer *m)
{
    if (!m)
        return NULL;

    CS_Streamer *s = (CS_Streamer *)av_malloc(sizeof(CS_Streamer));
    if (!s)
        return NULL;

    s->ele = m;
    s->power = &m->power;
    s->stream = ele_mixer_stream;

    return s;
}

ELE_Mixer *ele_create_mixer(const char *name)
{
    ELE_Mixer *m = (ELE_Mixer *)av_mallocz(sizeof(ELE_Mixer));
    if (!m)
        return NULL;
    if (name)
        strcpy(m->name, name);
    else
        strcpy(m->name, "Mixer");

    return m;
}

void ele_mixer_destory(ELE_Mixer **mp)
{
    if (!mp || !*mp)
        return;

    ELE_Mixer *m = *mp;
    if (m->buf)
        cs_samples_destory(&m->buf);
    while(m->size--)
        if (m->sources[m->size])
            av_free(m->sources[m->size]);
}