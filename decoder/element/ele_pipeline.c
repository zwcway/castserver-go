#include "./ele_pipeline.h"

int ele_pipeline_stream(ELE_Pipeline *p, CS_Samples *s)
{
    if (!p)
        return AVERROR(ENOMEM);

    CS_Streamer **ptr = p->eles, *se;
    uint64_t begin = GetTickCount64(), start, end;

    for (; ptr && (se = *ptr) && (*se->power); ptr++)
    {
        start = GetTickCount64();
        se->stream(se->ele, s);
        end = GetTickCount64();

        se->cost = (int)(end - start);
    }

    end = GetTickCount64();

    p->cost = (int)(end - begin);
    if (p->maxCost < p->cost)
        p->maxCost = p->cost;

    return 0;
}

int ele_pipeline_add(ELE_Pipeline *p, CS_Streamer *s)
{
    if (!p || p->size >= ELE_PIPELINE_STREAM_SIZE)
        return AVERROR(ENOMEM);

    p->eles[p->size++] = s;

    return 0;
}

CS_Streamer *ele_pipeline_create_streamer(void *ele, int *power, Func_Stream stream)
{
    CS_Streamer *s = (CS_Streamer *)av_mallocz(sizeof(CS_Streamer));
    if (!s)
        return NULL;

    s->ele = ele;
    s->power = power;
    s->stream = stream;

    return s;
}

ELE_Pipeline *ele_create_pipeline()
{
    ELE_Pipeline *p = (ELE_Pipeline *)av_mallocz(sizeof(ELE_Pipeline));

    if (!p)
        return AVERROR(ENOMEM);

    p->size = 0;
    p->eles[ELE_PIPELINE_STREAM_SIZE] = NULL;

    return p;
}

void ele_pipeline_destory(ELE_Pipeline **pp)
{
    if (!pp || !*pp)
        return;

    ELE_Pipeline *p = (ELE_Pipeline *)pp;

    if (p->buf)
        av_free(p->buf);

    while (p->size--)
        if (p->eles[p->size])
            av_free(p->eles[p->size]);

    *pp = NULL;
}