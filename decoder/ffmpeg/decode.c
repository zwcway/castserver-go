
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

static AVCodecContext *codecCtx;
static AVFormatContext *formatCtx;
static SwrContext *SwrCtx;
static AVFrame *avFrame = NULL;
static AVPacket *avPacket = NULL;
static int streamIndex;
static uint8_t *buffer = NULL;
static int bufferSize = 0;
static enum AVSampleFormat outputFmt;

static const int go_init_resample(int rate, int channels, enum AVSampleFormat fmt)
{
    if (codecCtx == NULL)
    {
        return -250;
    }
    int64_t inChannelLayout = av_get_default_channel_layout(codecCtx->channels);

    if (SwrCtx == NULL)
    {
        // 初始化转码器
        SwrCtx = swr_alloc();
        if (SwrCtx == NULL)
        {
            return -1;
        }
    }
    if (fmt == AV_SAMPLE_FMT_NONE)
    {
        fmt = codecCtx->sample_fmt;
    }
    outputFmt = fmt;

    SwrCtx = swr_alloc_set_opts(SwrCtx,
                                inChannelLayout, outputFmt, codecCtx->sample_rate,
                                inChannelLayout, codecCtx->sample_fmt, codecCtx->sample_rate,
                                0, NULL);

    int ret = swr_init(SwrCtx);
    if (ret < 0)
    {
        return ret;
    }

    return 0;
}

static void go_free()
{
    if (SwrCtx != NULL)
    {
        swr_close(SwrCtx);
        swr_free(&SwrCtx);
    }
    if (codecCtx != NULL)
    {
        avcodec_free_context(&codecCtx);
    }
    if (formatCtx != NULL)
    {
        avformat_close_input(&formatCtx);
    }

    if (avPacket != NULL)
    {
        av_packet_unref(avPacket);
        av_packet_free(&avPacket);
    }
    if (avFrame != NULL)
    {
        av_frame_free(&avFrame);
    }
    if (buffer != NULL)
    {
        av_freep(&buffer);
    }
}

static const int go_init(const char *fileName,
                         int *rate, int *channels, enum AVSampleFormat *fmt)
{
    // 注册所有解码器
    // av_register_all()

    // 初始化输入上下文
    int ret = avformat_open_input(&formatCtx, fileName, NULL, NULL);
    if (ret != 0)
    {
        go_free();
        return ret;
    }
    ret = avformat_find_stream_info(formatCtx, NULL);
    if (ret != 0)
    {
        go_free();
        return (ret);
    }

    // 查找音频流索引
    streamIndex = av_find_best_stream(formatCtx, AVMEDIA_TYPE_AUDIO, -1, -1, NULL, 0);

    if (streamIndex < 0)
    {
        go_free();
        return (streamIndex);
    }
    // 获取音频流
    const AVStream *audioStream = go_streams_index((const AVStream **)formatCtx->streams, streamIndex);
    if (audioStream == NULL)
    {
        go_free();
        return -255;
    }

    // 获取编码器
    AVCodec *audioCodec = avcodec_find_decoder(audioStream->codecpar->codec_id);
    if (audioCodec == NULL)
    {
        go_free();
        return -254;
    }

    // 获取编解码器上下文
    codecCtx = avcodec_alloc_context3(audioCodec);
    if (codecCtx == NULL)
    {
        go_free();
        return -253;
    }

    // 复制编解码器参数
    ret = avcodec_parameters_to_context(codecCtx, audioStream->codecpar);
    if (ret != 0)
    {
        go_free();
        return (ret);
    }

    if (codecCtx->channels > 16 || codecCtx->channels == 0)
    {
        go_free();
        return -252;
    }

    // 初始化解码器
    ret = avcodec_open2(codecCtx, audioCodec, NULL);
    if (ret < 0)
    {
        go_free();
        return (ret);
    }

    *channels = codecCtx->channels;
    *rate = codecCtx->sample_rate;
    *fmt = codecCtx->sample_fmt;

    avFrame = av_frame_alloc();
    if (avFrame == NULL)
    {
        go_free();
        return -1;
    }

    avPacket = av_packet_alloc();
    if (avPacket == NULL)
    {
        go_free();
        return -1;
    }

    return 0;
}

static const int go_seek(int p)
{
    const AVStream *stream = go_streams_index((const AVStream **)formatCtx->streams, streamIndex);
    const int64_t pos = go_seek_time(stream, p);
    return av_seek_frame(formatCtx, streamIndex, pos, AVSEEK_FLAG_BACKWARD | AVSEEK_FLAG_FRAME);
}

static const int go_decode(uint8_t **cBuffer, int *cSize)
{
    int ret = 0;
    int samples = 0;
    int is_eof = 0;
    while (1)
    {
        ret = av_read_frame(formatCtx, avPacket);
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
            ret = avcodec_send_packet(codecCtx, avPacket);
            if (ret == AVERROR(EAGAIN))
            {
                // 需要先receive
            }
            else if (ret < 0)
                return ret;
            else
                av_packet_unref(avPacket);
        }

        ret = avcodec_receive_frame(codecCtx, avFrame);
        if (ret >= 0)
        {
            samples = avFrame->nb_samples;
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

    int outBufferSize = av_samples_get_buffer_size(NULL, codecCtx->channels, samples, outputFmt, 0);

    if (bufferSize < outBufferSize)
    {
        if (buffer != NULL)
            av_freep(&buffer);

        buffer = (uint8_t *)av_malloc(outBufferSize);
        if (buffer == NULL)
        {
            *cSize = outBufferSize;
            bufferSize = 0;
            return -1;
        }
        bufferSize = outBufferSize;
    }
    *cBuffer = buffer;
    *cSize = bufferSize;

    ret = swr_convert(SwrCtx,
                      // buffer 是一维数组，因此初始化 swr 参数时，
                      // fmt 必须是 packed 类型，不可以是 panlar 类型,
                      // 否则转码多声道时会崩溃
                      &buffer, samples,
                      // 声道有可能超过8个，因此使用 extened_data
                      (const uint8_t **)avFrame->extended_data, samples);
    if (ret < 0)
        return ret;

    return ret;
}