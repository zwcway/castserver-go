
#ifndef ELE_MIXER
#define ELE_MIXER

#include "./samples.h"
#include "./pipeline.h"
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
     * @brief 输出的格式
     * 
     */
    CS_Format format;

    _CS_Mixer_Sourcer *sources[CS_MIXER_SOURCES_SIZE + 1];
    int size;

} ELE_Mixer;

ELE_Mixer *ele_create_mixer();
int ele_mixer_stream(void *mp, CS_Samples *s);
int cs_mixer_setOutputFormat(ELE_Mixer *m, CS_Format f);

#endif // ELE_MIXER