#include "../spaeker.h"
#include <unistd.h>
#include <libavformat/avformat.h>

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

    GOAVDecoder *ctx = NULL;
    int rate = 0;
    int channels = 0;
    enum AVSampleFormat fmt = 0;
    int ret = go_init(&ctx, fileName, &rate, &channels, &fmt);
    if (ret < 0)
    {
        av_strerror(ret, libav_errors, 200);
        goto _exit_;
    }

    ret = go_init_resample(ctx, rate, ctx->codecCtx->channel_layout, AV_SAMPLE_FMT_DBLP);
    if (ret < 0)
    {
        av_strerror(ret, libav_errors, 200);
        goto _exit_;
    }

    GOResample *rc = NULL;
    ret = go_swr_init(&rc, rate, ctx->codecCtx->channel_layout, fmt, rate, ctx->codecCtx->channel_layout, AV_SAMPLE_FMT_S16P);
    if (ret < 0)
    {
        av_strerror(ret, libav_errors, 200);
        goto _exit_;
    }

    // go_seek(ctx, 180 * 44100);
    while (1)
    {
        ret = go_decode(ctx);
        if (ret < 0)
        {
            av_strerror(ret, libav_errors, 200);
            goto _exit_;
        }

        rc->in_buffer = ctx->buffer;
        
        ret = go_convert(rc, ret);
        if (ret < 0) {
            av_strerror(ret, libav_errors, 200);
            goto _exit_;
        }

        printf("decoded samples count %d\n", ret);
    }

_exit_:
    printf("error %s\n", libav_errors);
    go_free(&ctx);
    exit(255);
}