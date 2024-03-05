#ifndef CS_SPEAKER
#define CS_SPEAKER

#include <soundio/soundio.h>
#include "samples.h"
#include "ele_mixer.h"

typedef struct CS_Speaker
{
    char name[16];
    
    struct SoundIo *io;
    enum SoundIoBackend backend;
    struct SoundIoDevice *device;
    struct SoundIoOutStream *outstream;
    char device_id[255];
    
    int pause;

    /**
     * @brief 输出格式
     * 
     */
    CS_Format fmt;
    CS_Samples *buf;
    ELE_Mixer *mixer;
    int samples_left;
} CS_Speaker;

CS_Speaker *cs_create_speaker();
void cs_speaker_destory(CS_Speaker **s);
int cs_speaker_setFormat(CS_Speaker *sp, CS_Format f);
#endif // CS_SPEAKER