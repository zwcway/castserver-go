
#include <stdio.h>
#include <libavformat/avformat.h>
#include <libswresample/swresample.h>

static const AVStream *go_streams_index(const AVStream **streams, int n)
{
    return streams[n];
}

static const int64_t go_seek_time(const AVStream *in, int seconds)
{
    int64_t startTime = seconds * AV_TIME_BASE;
    int64_t target_time = av_rescale_q(startTime, AV_TIME_BASE_Q, in->time_base);
    return target_time;
}

static const uint8_t go_get_array(const uint8_t *arr, int index)
{
    return arr[index];
}
static int go_averror_is_eof(int code)
{
    return code == AVERROR_EOF;
}

typedef struct
{
    AVCodecContext *codecCtx;
    AVFormatContext *formatCtx;
    SwrContext *swrCtx;
    AVFrame *avFrame;
    AVPacket *avPacket;
    int streamIndex;
    uint8_t *buffer;
    int bufferSize;
    enum AVSampleFormat outputFmt;

} GOAVDecoder;

static const int go_init_resample(GOAVDecoder *ctx, int rate, int channels, enum AVSampleFormat fmt)
{
    if (ctx == NULL || ctx->codecCtx == NULL)
    {
        return -250;
    }
    int64_t inChannelLayout = av_get_default_channel_layout(ctx->codecCtx->channels);

    if (ctx->swrCtx == NULL)
    {
        // 初始化转码器
        ctx->swrCtx = swr_alloc();
        if (ctx->swrCtx == NULL)
        {
            return -1;
        }
    }
    if (fmt == AV_SAMPLE_FMT_NONE)
    {
        fmt = ctx->codecCtx->sample_fmt;
    }
    ctx->outputFmt = fmt;

    ctx->swrCtx = swr_alloc_set_opts(ctx->swrCtx,
                                     inChannelLayout, ctx->outputFmt, ctx->codecCtx->sample_rate,
                                     inChannelLayout, ctx->codecCtx->sample_fmt, ctx->codecCtx->sample_rate,
                                     0, NULL);

    int ret = swr_init(ctx->swrCtx);
    if (ret < 0)
    { 
        swr_free(&ctx->swrCtx);
        return ret;
    }

    return 0;
}

static void go_free(GOAVDecoder **ctx)
{
    if (*ctx == NULL)
    {
        return;
    }

    if ((*ctx)->swrCtx != NULL)
    {
        swr_close((*ctx)->swrCtx);
        swr_free(&(*ctx)->swrCtx);
    }
    if ((*ctx)->codecCtx != NULL)
    {
        avcodec_free_context(&(*ctx)->codecCtx);
    }
    if ((*ctx)->formatCtx != NULL)
    {
        avformat_close_input(&(*ctx)->formatCtx);
    }

    if ((*ctx)->avPacket != NULL)
    {
        av_packet_unref((*ctx)->avPacket);
        av_packet_free(&(*ctx)->avPacket);
    }
    if ((*ctx)->avFrame != NULL)
    {
        av_frame_free(&(*ctx)->avFrame);
    }
    if ((*ctx)->buffer != NULL)
    {
        av_freep(&(*ctx)->buffer);
    }
    av_freep(ctx);
}

static const int go_init(GOAVDecoder **ctxp, const char *fileName,
                         int *rate, int *channels, enum AVSampleFormat *fmt)
{
    // 注册所有解码器
    // av_register_all()
    if (ctxp == NULL)
    {
        return -2;
    }
    GOAVDecoder *ctx = *ctxp = (GOAVDecoder *)av_malloc(sizeof(GOAVDecoder));
    if (ctx == NULL)
    {
        return -1;
    }

    // 初始化输入上下文
    int ret = avformat_open_input(&ctx->formatCtx, fileName, NULL, NULL);
    if (ret != 0)
    {
        go_free(&ctx);
        return ret;
    }
    ret = avformat_find_stream_info(ctx->formatCtx, NULL);
    if (ret != 0)
    {
        go_free(&ctx);
        return (ret);
    }

    // 查找音频流索引
    ctx->streamIndex = av_find_best_stream(ctx->formatCtx, AVMEDIA_TYPE_AUDIO, -1, -1, NULL, 0);

    if (ctx->streamIndex < 0)
    {
        go_free(&ctx);
        return (ctx->streamIndex);
    }
    // 获取音频流
    const AVStream *audioStream = go_streams_index((const AVStream **)ctx->formatCtx->streams, ctx->streamIndex);
    if (audioStream == NULL)
    {
        go_free(&ctx);
        return -255;
    }

    // 获取编码器
    AVCodec *audioCodec = avcodec_find_decoder(audioStream->codecpar->codec_id);
    if (audioCodec == NULL)
    {
        go_free(&ctx);
        return -254;
    }

    // 获取编解码器上下文
    ctx->codecCtx = avcodec_alloc_context3(audioCodec);
    if (ctx->codecCtx == NULL)
    {
        go_free(&ctx);
        return -253;
    }

    // 复制编解码器参数
    ret = avcodec_parameters_to_context(ctx->codecCtx, audioStream->codecpar);
    if (ret != 0)
    {
        go_free(&ctx);
        return (ret);
    }

    if (ctx->codecCtx->channels > 16 || ctx->codecCtx->channels == 0)
    {
        go_free(&ctx);
        return -252;
    }

    // 初始化解码器
    ret = avcodec_open2(ctx->codecCtx, audioCodec, NULL);
    if (ret < 0)
    {
        go_free(&ctx);
        return (ret);
    }

    *channels = ctx->codecCtx->channels;
    *rate = ctx->codecCtx->sample_rate;
    *fmt = ctx->codecCtx->sample_fmt;

    ctx->avFrame = av_frame_alloc();
    if (ctx->avFrame == NULL)
    {
        go_free(&ctx);
        return -1;
    }

    ctx->avPacket = av_packet_alloc();
    if (ctx->avPacket == NULL)
    {
        go_free(&ctx);
        return -1;
    }

    return 0;
}

static const int go_seek(GOAVDecoder *ctx, int p)
{
    if (ctx == NULL)
    {
        return -2;
    }
    const AVStream *stream = go_streams_index((const AVStream **)ctx->formatCtx->streams, ctx->streamIndex);
    const int64_t pos = go_seek_time(stream, p);
    return av_seek_frame(ctx->formatCtx, ctx->streamIndex, pos, AVSEEK_FLAG_BACKWARD | AVSEEK_FLAG_FRAME);
}

static const int go_decode(GOAVDecoder *ctx)
{
    int ret = 0;
    int samples = 0;
    int is_eof = 0;

    if (ctx == NULL)
    {
        return -2;
    }

    while (1)
    {
        ret = av_read_frame(ctx->formatCtx, ctx->avPacket);
        if (ret == AVERROR_EOF)
        {
            is_eof = 1;
            // 解码完成，交给 avcodec_receive_frame 判断是否读取完成
        }
        else if (ret < 0)
            return ret;
        else
        {
            // 发送至解码队列
            // packet中可能包含多帧音频,需要多次读取
            ret = avcodec_send_packet(ctx->codecCtx, ctx->avPacket);
            if (ret == AVERROR(EAGAIN))
            {
                // 需要先receive
            }
            else if (ret < 0)
                return ret;
            else
                av_packet_unref(ctx->avPacket);
        }

        ret = avcodec_receive_frame(ctx->codecCtx, ctx->avFrame);
        if (ret >= 0)
        {
            samples = ctx->avFrame->nb_samples;
            break;
        }
        if (ret == AVERROR(EAGAIN))
        {
            if (is_eof)
                return AVERROR_EOF;
            // 继续读取
        }
        else if (ret < 0)
            return ret;
    }

    int outBufferSize = av_samples_get_buffer_size(NULL, ctx->codecCtx->channels, samples, ctx->outputFmt, 0);

    if (ctx->bufferSize < outBufferSize)
    {
        if (ctx->buffer != NULL)
            av_freep(&ctx->buffer);

        ctx->buffer = (uint8_t *)av_malloc(outBufferSize);
        if (ctx->buffer == NULL)
        {
            ctx->bufferSize = 0;
            return -1;
        }
        ctx->bufferSize = outBufferSize;
    }

    ret = swr_convert(ctx->swrCtx,
                      // buffer 是一维数组，因此初始化 swr 参数时，
                      // fmt 必须是 packed 类型，不可以是 panlar 类型,
                      // 否则转码多声道时会崩溃
                      &ctx->buffer, samples,
                      // 声道有可能超过8个，因此使用 extened_data
                      (const uint8_t **)ctx->avFrame->extended_data, samples);
    if (ret < 0)
        return ret;

    return ret;
}