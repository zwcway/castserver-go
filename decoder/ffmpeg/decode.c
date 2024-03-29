#include <stdio.h>
#include <libavformat/avformat.h>
#include <libswresample/swresample.h>
#include "resample/resample.c"

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

#define CHANNEL_MAX 32
typedef struct GOAVDecoder
{
    AVCodecContext *codecCtx;
    AVFormatContext *formatCtx;
    GOResample *swrCtx;
    AVFrame *avFrame;
    AVPacket *avPacket;
    AVStream *stream;
    int streamIndex;
    int duration;
    /**
     * @brief 指向 GOResample 中的 out_buffer
     * 
     */
    uint8_t *buffer;
    /**
     * @brief 一次解码的样本数量
     * 
     */
    int nb_samples;
    enum AVSampleFormat outputFmt;
    int debug;
    int finished;
    int64_t start_pts;
    AVRational start_pts_tb;
    int64_t next_pts;
    AVRational next_pts_tb;

    /**
     * @brief 声道映射关系
     * 
     */
    int channel_index[CHANNEL_MAX];
} GOAVDecoder;

static const int go_init_resample(GOAVDecoder *ctx, int rate, int64_t channel_layout, enum AVSampleFormat fmt)
{
    int ret = 0;

    if (ctx == NULL || ctx->codecCtx == NULL)
    {
        return -250;
    }

    if (fmt == AV_SAMPLE_FMT_NONE)
    {
        fmt = ctx->codecCtx->sample_fmt;
    }
    ctx->outputFmt = fmt;

    // 初始化转码器
    ret = go_swr_init(&ctx->swrCtx,
                        ctx->codecCtx->sample_rate, ctx->codecCtx->channel_layout, ctx->codecCtx->sample_fmt,
                        rate, channel_layout, fmt);

    return ret;
}

static void go_free(GOAVDecoder **ctx)
{
    if (*ctx == NULL)
    {
        return;
    }

    if ((*ctx)->swrCtx != NULL)
    {
        go_swr_free(&(*ctx)->swrCtx);
    }

    if ((*ctx)->codecCtx != NULL)
    {
        avcodec_flush_buffers((*ctx)->codecCtx);
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
    (*ctx)->stream = NULL;
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
    GOAVDecoder *ctx = *ctxp = (GOAVDecoder *)av_mallocz(sizeof(GOAVDecoder));
    if (ctx == NULL)
    {
        return -1;
    }

    // 初始化输入上下文
    int ret = avformat_open_input(&ctx->formatCtx, fileName, NULL, NULL);
    if (ret != 0)
        goto _exit_;

    ret = avformat_find_stream_info(ctx->formatCtx, NULL);
    if (ret != 0)
        goto _exit_;

    // 查找音频流索引
    ctx->streamIndex = av_find_best_stream(ctx->formatCtx, AVMEDIA_TYPE_AUDIO, -1, -1, NULL, 0);

    if (ctx->streamIndex < 0)
    {
        ret = ctx->streamIndex;
        goto _exit_;
    }

    // 获取音频流
    ctx->stream = (AVStream *)go_streams_index((const AVStream **)ctx->formatCtx->streams, ctx->streamIndex);
    if (ctx->stream == NULL)
    {
        ret = -255;
        goto _exit_;
    }
    ctx->stream->discard = AVDISCARD_DEFAULT;

    // 获取编码器
    AVCodec *audioCodec = avcodec_find_decoder(ctx->stream->codecpar->codec_id);
    if (audioCodec == NULL)
    {
        ret = -254;
        goto _exit_;
    }

    // 获取编解码器上下文
    ctx->codecCtx = avcodec_alloc_context3(audioCodec);
    if (ctx->codecCtx == NULL)
    {
        ret = -253;
        goto _exit_;
    }

    // 复制编解码器参数
    ret = avcodec_parameters_to_context(ctx->codecCtx, ctx->stream->codecpar);
    if (ret != 0)
    {
        goto _exit_;
    }

    if (ctx->codecCtx->channels > 16 || ctx->codecCtx->channels == 0)
    {
        ret = -252;
        goto _exit_;
    }

    ctx->codecCtx->pkt_timebase = ctx->stream->time_base;
    ctx->codecCtx->codec_id = audioCodec->id;
    ctx->codecCtx->lowres = 0;
    ctx->start_pts = AV_NOPTS_VALUE;
    if ((ctx->formatCtx->iformat->flags & (AVFMT_NOBINSEARCH | AVFMT_NOGENSEARCH | AVFMT_NO_BYTE_SEEK)) && !ctx->formatCtx->iformat->read_seek)
    {
        ctx->start_pts = ctx->stream->start_time;
        ctx->start_pts_tb = ctx->stream->time_base;
    }

    // 初始化解码器
    ret = avcodec_open2(ctx->codecCtx, audioCodec, NULL);
    if (ret < 0)
    {
        goto _exit_;
    }

    if (ctx->codecCtx->channel_layout == 0) {
        ctx->codecCtx->channel_layout = av_get_default_channel_layout(ctx->codecCtx->channels);
    }
    
    for (ret = 0; ret < CHANNEL_MAX; ret ++)
        ctx->channel_index[ret] = -1;
    
    for (ret = 0; ret < ctx->codecCtx->channels; ret ++)
        ctx->channel_index[ret] = av_channel_layout_extract_channel(ctx->codecCtx->channel_layout, ret);
    
    *channels = ctx->codecCtx->channels;
    *rate = ctx->codecCtx->sample_rate;
    *fmt = ctx->codecCtx->sample_fmt;

    ctx->avFrame = av_frame_alloc();
    if (ctx->avFrame == NULL)
    {
        ret = -1;
        goto _exit_;
    }

    ctx->avPacket = av_packet_alloc();
    if (ctx->avPacket == NULL)
    {
        ret = -1;
        goto _exit_;
    }

    return 0;

_exit_:
    go_free(ctxp);
    return ret;
}

static int is_realtime(AVFormatContext *s)
{
    if (!strcmp(s->iformat->name, "rtp") || !strcmp(s->iformat->name, "rtsp") || !strcmp(s->iformat->name, "sdp"))
        return 1;

    if (s->pb && (!strncmp(s->url, "rtp:", 4) || !strncmp(s->url, "udp:", 4)))
        return 1;
    return 0;
}

/**
 * @brief
 *
 * @param ctx
 * @param p 新的位置（秒）
 * @return const int
 */
static const int go_seek(GOAVDecoder *ctx, int p)
{
    if (ctx == NULL || ctx->stream == NULL)
        return -2;

    int flags = AVSEEK_FLAG_FRAME;
    avcodec_flush_buffers(ctx->codecCtx);
    int64_t pos = go_seek_time(ctx->stream, p);

    if (ctx->formatCtx->start_time != AV_NOPTS_VALUE)
        pos += ctx->formatCtx->start_time;

    int ret = avformat_seek_file(ctx->formatCtx, ctx->streamIndex, INT64_MIN, pos, INT64_MAX, 0);

    if (ret >= 0)
    {
        ctx->finished = 0;
    }

    return ret;
}

static const int go_decode(GOAVDecoder *ctx)
{
    int ret = 0;

    if (ctx == NULL)
    {
        return -2;
    }

    while (1)
    {
    _receive_:
        ret = avcodec_receive_frame(ctx->codecCtx, ctx->avFrame);
        if (ret >= 0)
            break; // 交给转码
        if (ret == AVERROR(EAGAIN))
        {
            // 继续解码
        }
        else
            goto _flush_buffer_;

        if (ctx->finished)
        { // 已经解码完成，无需再次解码
            ret = AVERROR_EOF;
            goto _flush_buffer_;
        }

        while (1)
        {
            ret = av_read_frame(ctx->formatCtx, ctx->avPacket);
            if (ret == AVERROR_EOF)
            {
                ctx->finished = 1;
                // 解码完成，清空已缓存的帧
                avcodec_send_packet(ctx->codecCtx, NULL);
                goto _receive_;
            }
            else if (ret < 0)
                goto _flush_buffer_;

            if (ctx->avPacket->stream_index != ctx->streamIndex)
            {
                av_packet_unref(ctx->avPacket);
                continue;
            }
            break;
        }
        // 发送至解码队列
        // packet中可能包含多帧音频,需要多次读取
        ret = avcodec_send_packet(ctx->codecCtx, ctx->avPacket);
        if (ret == AVERROR(EAGAIN))
            // 需要先receive
            continue;

        av_packet_unref(ctx->avPacket);
        if (ret < 0)
            goto _flush_buffer_;
    }

    ctx->swrCtx->in_buffer = (uint8_t*)ctx->avFrame->extended_data;

    ctx->nb_samples = ctx->avFrame->nb_samples;

    ret = go_convert(ctx->swrCtx, ctx->avFrame->nb_samples);

    ctx->swrCtx->in_buffer = NULL;
    ctx->buffer = ctx->swrCtx->out_buffer;
    if (ret < 0)
        return ret;

    // 当面解码帧的时间
    ctx->duration = ctx->avFrame->pts * av_q2d(ctx->stream->time_base);

    return ret;
_flush_buffer_:
    ctx->finished = 1;
    avcodec_flush_buffers(ctx->codecCtx);
    return ret;
}