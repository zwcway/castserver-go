#include <math.h>
#include "./ele_volume.h"

int ele_volume_stream(void *ele, CS_Samples *s)
{
    int c, i;

    if (!ele)
        return AVERROR(ENOMEM);

    ELE_Volume *v = (ELE_Volume *)ele;

    if (!v->power || v->gain == 1.0f)
        return 0;

    for (c = 0; c < s->format.chs; c++)
        for (i = 0; i < s->last_nb_samples; i++)
            s->data[c][i] *= v->gain;
}

int ele_volume_on(void *ele)
{
    if (!ele)
        return AVERROR(ENOMEM);
    ((ELE_Volume *)ele)->power = 1;
}

int ele_volume_off(void *ele)
{
    if (!ele)
        return AVERROR(ENOMEM);
    ((ELE_Volume *)ele)->power = 0;
}

int ele_volume_isOn(void *ele)
{
    if (!ele)
        return AVERROR(ENOMEM);
    return ((ELE_Volume *)ele)->power;
}

int ele_volume_setMute(void *ele, int b)
{
    if (!ele)
        return AVERROR(ENOMEM);
    ELE_Volume *v = (ELE_Volume *)ele;
    v->mute = b;
    ele_volume_setVolume(ele, v->volume);
}

int ele_volume_isMute(void *ele)
{
    if (!ele)
        return AVERROR(ENOMEM);
    ELE_Volume *v = (ELE_Volume *)ele;
    return v->mute;
}

int ele_volume_setVolume(void *ele, double vol)
{
    if (!ele)
        return AVERROR(ENOMEM);
    ELE_Volume *v = (ELE_Volume *)ele;

    v->volume = vol;

    if (v->mute || v->volume == 0.0f)
        v->gain = 0;
    else if (v->base == 1.0f)
        v->gain = v->volume;
    else
        v->gain = pow(v->base, v->volume);

    return 0;
}

double ele_volume_volume(void *ele)
{
    if (!ele)
        return AVERROR(ENOMEM);
    ELE_Volume *v = (ELE_Volume *)ele;

    return v->volume;
}

ELE_Volume *ele_create_volume(double vol)
{
    ELE_Volume *v = (ELE_Volume *)av_mallocz(sizeof(ELE_Volume));
    if (!v)
        return NULL;

    strcpy(v->name, "Volume");
    v->base = 1.0f;
    v->power = 1;
    v->volume = vol;

    ele_volume_setVolume(v, vol);

    return v;
}
