#include "../speaker.h"
#include <unistd.h>
#include <libavformat/avformat.h>
#include "../ele_decode.c"
#include "../ele_mixer.c"

static char libav_errors[256] = {0};

int main(int argc, char *argv[])
{
    char fileName[256] = {0};
    int opt;
    while ((opt = getopt(argc, argv, "f:")) != -1)
    {
        switch (opt)
        {
        case 'f':
            strncpy(fileName, optarg, 255);
            break;
        }
    }
    if (fileName[0] == 0)
    {
        printf("Usage: -f filename\n");
        exit(1);
    }
    // if (!access(fileName, 0))
    // {
    //     printf("file not exists: %s\n", fileName);
    //     exit(1);
    // }

    ELE_Decoder *decoder = NULL;
    ELE_Mixer *mixer = NULL;
    ELE_Pipeline *pipeline = NULL;

    CS_Speaker *speaker = NULL;
    int err;
    CS_Format format;

    if (!(decoder = ele_create_decoder()))
        goto _exit_;

    if (!(mixer = ele_create_mixer(NULL)))
        goto _exit_;

    if (!(pipeline = ele_create_pipeline()))
        goto _exit_;

    if (!(speaker = cs_create_speaker()))
        goto _exit_;

    ele_mixer_add(mixer, ele_decoder_sourcer(decoder));
    ele_pipeline_add(pipeline, ele_mixer_streamer(mixer));

    if (err = ele_decoder_open(decoder, fileName))
        goto _exit_;


    ele_decoder_audioFormat(decoder, &format);
    printf("[%d/%dbit/%dch]%s\n", format.srate, format.bitw * 8, format.chs, fileName);


_exit_:
    if (err)
    {
        av_strerror(err, libav_errors, 200);
        printf("error %s\n", libav_errors);
    }
    if (speaker)
        cs_speaker_destory(&speaker);
    if (pipeline) 
        ele_pipeline_destory(&pipeline);
    if (mixer)
        ele_mixer_destory(&mixer);
    if (decoder)
        ele_decoder_destory(&decoder);

    exit(255);
}