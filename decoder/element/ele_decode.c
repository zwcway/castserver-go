#include <stdio.h>
#include <libavformat/avformat.h>
#include <libswresample/swresample.h>
#include "./samples.h"
#include "./ele_mixer.h"

static const int64_t _ele_decoder_seek_time(const AVStream *in, int seconds)
{
    int64_t startTime = seconds * AV_TIME_BASE;
    int64_t target_time = av_rescale_q(startTime, AV_TIME_BASE_Q, in->time_base);
    return target_time;
}

typedef struct ELE_Decoder
{
    AVCodecContext *codecCtx;
    AVFormatContext *formatCtx;

    AVFrame *avFrame;
    AVPacket *avPacket;
    AVStream *stream;
    int streamIndex;
    int duration;
    /**
     * @brief 剩余的样本数量，待取出
     *
     */
    int left_samples;
    int is_planar;
    int debug;
    int finished;
    int64_t start_pts;
    AVRational start_pts_tb;
    int64_t next_pts;
    AVRational next_pts_tb;

    CS_Format fmt;
    /**
     * @brief 声道映射关系
     *
     */
    int channel_index[CHANNEL_MAX];
} ELE_Decoder;

void _ele_decoder_free(ELE_Decoder *ctx)
{
    if (ctx == NULL)
        return;

    if ((ctx)->avPacket != NULL)
    {
        av_packet_unref((ctx)->avPacket);
        av_packet_free(&(ctx)->avPacket);
    }
    if ((ctx)->avFrame != NULL)
    {
        av_frame_free(&(ctx)->avFrame);
    }
    if ((ctx)->codecCtx != NULL)
    {
        avcodec_flush_buffers((ctx)->codecCtx);
        avcodec_free_context(&(ctx)->codecCtx);
    }
    if ((ctx)->formatCtx != NULL)
    {
        avformat_close_input(&(ctx)->formatCtx);
    }
    (ctx)->stream = NULL;
}

void ele_decoder_destory(ELE_Decoder **ctx)
{
    if (!ctx || !*ctx)
        return;

    _ele_decoder_free(*ctx);
    av_freep(ctx);
}

ELE_Decoder *ele_create_decoder()
{
    ELE_Decoder *ctx = (ELE_Decoder *)av_mallocz(sizeof(ELE_Decoder));
    if (ctx == NULL)
    {
        return NULL;
    }

    return ctx;
}

int ele_decoder_open(ELE_Decoder *ctx, const char *fileName)
{
    // 注册所有解码器
    // av_register_all()
    if (ctx == NULL)
    {
        return -2;
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
    ctx->stream = ctx->formatCtx->streams[ctx->streamIndex];
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

    if (ctx->codecCtx->channel_layout == 0)
    {
        ctx->codecCtx->channel_layout = av_get_default_channel_layout(ctx->codecCtx->channels);
    }

    for (ret = 0; ret < CHANNEL_MAX; ret++)
        ctx->channel_index[ret] = -1;

    for (ret = 0; ret < ctx->codecCtx->channels; ret++)
        ctx->channel_index[ret] = av_channel_layout_extract_channel(ctx->codecCtx->channel_layout, ret);

    ctx->fmt.bit = av_get_bytes_per_sample(ctx->codecCtx->sample_fmt);
    ctx->fmt.chs = ctx->codecCtx->channels;
    ctx->fmt.srate = ctx->codecCtx->sample_rate;

    ctx->is_planar = av_get_planar_sample_fmt(ctx->codecCtx->sample_fmt) == ctx->codecCtx->sample_fmt;

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
    _ele_decoder_free(ctx);
    return ret;
}

/**
 * @brief
 * 实现接口 @see Func_AudioFormat
 *
 * @param e
 * @param f
 */
int ele_decoder_audioFormat(void *e, CS_Format *f)
{
    if (!e || !f)
        return 1;

    *f = ((ELE_Decoder *)e)->fmt;

    return 0;
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
static const int ele_decoder_seek(ELE_Decoder *ctx, int p)
{
    if (ctx == NULL || ctx->stream == NULL)
        return -2;

    int flags = AVSEEK_FLAG_FRAME;
    avcodec_flush_buffers(ctx->codecCtx);
    int64_t pos = _ele_decoder_seek_time(ctx->stream, p);

    if (ctx->formatCtx->start_time != AV_NOPTS_VALUE)
        pos += ctx->formatCtx->start_time;

    int ret = avformat_seek_file(ctx->formatCtx, ctx->streamIndex, INT64_MIN, pos, INT64_MAX, 0);

    if (ret >= 0)
    {
        ctx->finished = 0;
    }

    return ret;
}

int _ele_decoder_decode(ELE_Decoder *ctx)
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

    // 当面解码帧的时间
    ctx->duration = ctx->avFrame->pts * av_q2d(ctx->stream->time_base);

    return ret;
_flush_buffer_:
    ctx->finished = 1;
    avcodec_flush_buffers(ctx->codecCtx);
    return ret;
}

int ele_decoder_stream(void *ctx, CS_Samples *s)
{
    int need = s->req_nb_samples;
    int ret, c, ch;
    ELE_Decoder *d = (ELE_Decoder *)ctx;

    while (need > 0)
    {
        if (d->left_samples == 0)
        {
            // 继续解码
            if (ret = _ele_decoder_decode(d))
            {
                return ret;
            }
            d->left_samples = d->avFrame->nb_samples;
        }

        if (d->left_samples > 0)
        {
            c = need > d->left_samples ? d->left_samples : need;
            int copy_size = c * d->fmt.bit;
            int dof = (s->req_nb_samples - need) * d->fmt.bit;

            if (d->is_planar)
            {
                // planar 格式：声道1所有样本 + 声道2所有样本 ...

                int sof = (d->avFrame->nb_samples * d->fmt.bit) - d->left_samples * d->fmt.bit;

                for (ch = 0; ch < d->avFrame->channels && ch < s->format.chs; ch++)
                    memcpy(s->raw_data[ch] + dof, d->avFrame->extended_data[ch] + sof, copy_size);
            }
            else
            {
                // packed 格式：声道1样本1 + 声道2样本1 ... 声道1样本2 + 声道2样本2 ...
                int offset = (d->avFrame->nb_samples - d->left_samples) * d->fmt.bit * d->avFrame->channels;
                uint8_t *sd = d->avFrame->extended_data[0] + offset;
                int ch_width = d->fmt.bit * d->avFrame->channels;

                switch (d->fmt.bit)
                {
                case 1:
                    for (int i = 0; i < copy_size; i += d->fmt.bit)
                        ((uint8_t *)(s->raw_data[ch] + dof))[i] = ((uint8_t *)(sd + ch_width * i))[0];
                    break;
                case 2:
                    for (int i = 0; i < copy_size; i += d->fmt.bit)
                        ((uint16_t *)(s->raw_data[ch] + dof))[i] = ((uint16_t *)(sd + ch_width * i))[0];
                    break;
                case 4:
                    for (int i = 0; i < copy_size; i += d->fmt.bit)
                        ((uint32_t *)(s->raw_data[ch] + dof))[i] = ((uint32_t *)(sd + ch_width * i))[0];
                    break;
                case 8:
                    for (int i = 0; i < copy_size; i += d->fmt.bit)
                        ((uint64_t *)(s->raw_data[ch] + dof))[i] = ((uint64_t *)(sd + ch_width * i))[0];
                    break;
                }
            }

            need -= c;
            d->left_samples -= c;
            continue;
        }
        break;
    }

    return 0;
_exit_:

    s->last_nb_samples = s->req_nb_samples - need;
#if 0
    // 空余的空间静音
    if (need > 0) {
        int size = need * ctx->fmt.bit;
        int offset = (s->req_nb_samples) * ctx->fmt.bit - size;
        for (ch = 0; ch < s->format.chs; ch ++) 
            memset(s->raw_data[ch] + offset, 0, size);
    }
#endif
}

CS_Sourcer *ele_decoder_sourcer(void *ele)
{
    return cs_create_sourcer(ele, ele_decoder_stream, ele_decoder_audioFormat);
}