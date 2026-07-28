
#ifndef ELE_MIXER
#define ELE_MIXER

#include "./samples.h"
#include "./ele_pipeline.h"
#include "./resample.h"

typedef struct CS_Sourcer
{
    /**
     * @brief
     *
     */
    void *ele;
    Func_Stream f_stream;
    Func_AudioFormat f_format;
} CS_Sourcer;

inline CS_Sourcer *cs_create_sourcer(void *ele, Func_Stream fs, Func_AudioFormat ff)
{
    CS_Sourcer *s = (CS_Sourcer *)av_malloc(sizeof(CS_Sourcer));
    if (!s)
        return NULL;
    s->ele = ele;
    s->f_format = ff;
    s->f_stream = fs;
    return s;
}
typedef struct _CS_Mixer_Sourcer
{
    CS_Sourcer *cs;

    /**
     * @brief 转码器
     *
     */
    CS_Resample *rs;

} _CS_Mixer_Sourcer;

#define CS_MIXER_SOURCES_SIZE 15
typedef struct ELE_Mixer
{
    char name[16];
    int power;

    CS_Samples *buf;
    /**
     * @brief 输出的格式，与 buf->format 相同 
     *
     */
    CS_Format format;

    _CS_Mixer_Sourcer *sources[CS_MIXER_SOURCES_SIZE + 1];
    int size;

} ELE_Mixer;

ELE_Mixer *ele_create_mixer(const char *);
void ele_mixer_destory(ELE_Mixer **);
CS_Streamer *ele_mixer_streamer(ELE_Mixer *);
int ele_mixer_add(ELE_Mixer *m, CS_Sourcer *src);
int ele_mixer_stream(void *mp, CS_Samples *s);
int ele_mixer_setOutputFormat(ELE_Mixer *m, CS_Format f);

#endif // ELE_MIXER