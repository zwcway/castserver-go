#include "speaker.h"

static const enum SoundIoFormat _cs_speaker_format(CS_Bits bit)
{
    switch (bit)
    {
    case Bits_S8:
        return SoundIoFormatS8;
    case Bits_U8:
        return SoundIoFormatU8;
    case Bits_S16LE:
        return SoundIoFormatS16LE;
    case Bits_U16LE:
        return SoundIoFormatU16LE;
    case Bits_S24LE:
        return SoundIoFormatS24LE;
    case Bits_U24LE:
        return SoundIoFormatU24LE;
    case Bits_S32LE:
        return SoundIoFormatS32LE;
    case Bits_U32LE:
        return SoundIoFormatU32LE;
    case Bits_32LEF:
        return SoundIoFormatFloat32LE;
    case Bits_64LEF:
        return SoundIoFormatFloat64LE;
    default:
        return SoundIoFormatInvalid;
    }
}

static const struct SoundIoChannelLayout *_cs_speaker_layout(CS_LayoutMask layout)
{
    enum SoundIoChannelLayoutId l;

    switch (layout)
    {
    case LayoutMask_Mono:
        l = SoundIoChannelLayoutIdMono;
        break;
    default:
    case LayoutMask_Stereo:
        l = SoundIoChannelLayoutIdStereo;
        break;
    case LayoutMask_21:
        l = SoundIoChannelLayoutId2Point1;
        break;
    case LayoutMask_30:
        l = SoundIoChannelLayoutId3Point0;
        break;
    case AV_CH_LAYOUT_2_1:
        l = SoundIoChannelLayoutId3Point0Back;
        break;
    case AV_CH_LAYOUT_3POINT1:
        l = SoundIoChannelLayoutId3Point1;
        break;
    case AV_CH_LAYOUT_4POINT0:
        l = SoundIoChannelLayoutId4Point0;
        break;
    case AV_CH_LAYOUT_QUAD:
        l = SoundIoChannelLayoutIdQuad;
        break;
    case AV_CH_LAYOUT_2_2:
        l = SoundIoChannelLayoutIdQuadSide;
        break;
    case AV_CH_LAYOUT_4POINT1:
        l = SoundIoChannelLayoutId4Point1;
        break;
    case AV_CH_LAYOUT_5POINT0_BACK:
        l = SoundIoChannelLayoutId5Point0Back;
        break;
    case AV_CH_LAYOUT_5POINT0:
        l = SoundIoChannelLayoutId5Point0Side;
        break;
    case AV_CH_LAYOUT_5POINT1:
        l = SoundIoChannelLayoutId5Point1;
        break;
    case AV_CH_LAYOUT_5POINT1_BACK:
        l = SoundIoChannelLayoutId5Point1Back;
        break;
    case AV_CH_LAYOUT_6POINT0:
        l = SoundIoChannelLayoutId6Point0Side;
        break;
    case AV_CH_LAYOUT_6POINT0_FRONT:
        l = SoundIoChannelLayoutIdHexagonal;
        break;
    case AV_CH_LAYOUT_6POINT1:
        l = SoundIoChannelLayoutId6Point1;
        break;
    case AV_CH_LAYOUT_6POINT1_BACK:
        l = SoundIoChannelLayoutId6Point1Back;
        break;
    case AV_CH_LAYOUT_6POINT1_FRONT:
        l = SoundIoChannelLayoutId6Point1Front;
        break;
    case AV_CH_LAYOUT_7POINT0:
        l = SoundIoChannelLayoutId7Point0;
        break;
    case AV_CH_LAYOUT_7POINT0_FRONT:
        l = SoundIoChannelLayoutId7Point0Front;
        break;
    case AV_CH_LAYOUT_7POINT1:
        l = SoundIoChannelLayoutId7Point1;
        break;
    case AV_CH_LAYOUT_7POINT1_WIDE:
        l = SoundIoChannelLayoutId7Point1Wide;
        break;
    case AV_CH_LAYOUT_7POINT1_WIDE_BACK:
        l = SoundIoChannelLayoutId7Point1WideBack;
        break;
    case AV_CH_LAYOUT_OCTAGONAL:
        l = SoundIoChannelLayoutIdOctagonal;
        break;
    }

    return soundio_channel_layout_get_builtin(l);
}

int _cs_speaker_index_by_id(CS_Speaker *sp, char *id)
{
    if (id)
    {
        int device_count = soundio_output_device_count(sp->io);
        for (int i = 0; i < device_count; i += 1)
        {
            struct SoundIoDevice *device = soundio_get_output_device(sp->io, i);
            bool select_this_one = strcmp(device->id, id) == 0 && device->is_raw == 1;
            soundio_device_unref(device);
            if (select_this_one)
            {
                return i;
            }
        }
        return -1;
    }

    int index = soundio_default_output_device_index(sp->io);
    return index;
}

static void write_callback(struct SoundIoOutStream *outstream, int frame_count_min, int frame_count_max)
{
    struct SoundIoChannelArea *areas;
    int err;
    int frames_left = frame_count_max;
    CS_Speaker *sp = (CS_Speaker *)outstream->userdata;
    struct SoundIoChannelArea *dst;
    uint8_t *src;
    int offset = (sp->buf->req_nb_samples - sp->samples_left) * sp->buf->format.bitw;
    int channel, i;

    for (;;)
    {
        int frame_count = frames_left;
        if ((err = soundio_outstream_begin_write(outstream, &areas, &frame_count)))
        {
            fprintf(stderr, "unrecoverable stream error: %s\n", soundio_strerror(err));
            exit(1);
        }
        if (!frame_count)
            break;
        const struct SoundIoChannelLayout *layout = &outstream->layout;

        frames_left -= frame_count;

        while (frame_count > 0)
        {
            if (sp->samples_left <= 0)
            {
                ele_mixer_stream(sp->mixer, sp->buf);
                sp->samples_left = sp->buf->last_nb_samples;
            }
            int copy_count = sp->samples_left > frame_count ? frame_count : sp->samples_left;

            if (sp->samples_left > 0)
            {
                sp->samples_left -= copy_count;
                frame_count -= copy_count;

                offset = (sp->buf->last_nb_samples - sp->samples_left) * sp->buf->format.bitw;

                switch (sp->buf->format.bitw)
                {
                case 1:
                    for (channel = 0; channel < layout->channel_count; channel++)
                    {
                        dst = &areas[channel];
                        src = (uint8_t *)(sp->buf->raw_data[channel]) + offset;
                        for (i = 0; i < copy_count; i++, dst->ptr += dst->step, src += sp->buf->format.bitw)
                            *((uint8_t *)dst->ptr) = *((uint8_t *)src);
                    }
                    break;
                case 2:
                    for (channel = 0; channel < layout->channel_count; channel++)
                    {
                        dst = &areas[channel];
                        src = (uint8_t *)(sp->buf->raw_data[channel]) + offset;
                        for (i = 0; i < copy_count; i++, dst->ptr += dst->step, src += sp->buf->format.bitw)
                            *((uint16_t *)dst->ptr) = *((uint16_t *)src);
                    }
                    break;
                case 3:
                    for (channel = 0; channel < layout->channel_count; channel++)
                    {
                        dst = &areas[channel];
                        src = (uint8_t *)(sp->buf->raw_data[channel]) + offset;
                        for (i = 0; i < copy_count; i++, dst->ptr += dst->step, src += sp->buf->format.bitw)
                        {
                            *(((uint8_t *)dst->ptr) + 0) = *(((uint8_t *)src) + 0);
                            *(((uint8_t *)dst->ptr) + 1) = *(((uint8_t *)src) + 1);
                            *(((uint8_t *)dst->ptr) + 2) = *(((uint8_t *)src) + 2);
                        }
                    }
                    break;
                case 4:
                    for (channel = 0; channel < layout->channel_count; channel++)
                    {
                        dst = &areas[channel];
                        src = (uint8_t *)(sp->buf->raw_data[channel]) + offset;
                        for (i = 0; i < copy_count; i++, dst->ptr += dst->step, src += sp->buf->format.bitw)
                            *((uint32_t *)dst->ptr) = *((uint32_t *)src);
                    }
                    break;
                case 8:
                    for (channel = 0; channel < layout->channel_count; channel++)
                    {
                        dst = &areas[channel];
                        src = (uint8_t *)(sp->buf->raw_data[channel]) + offset;
                        for (i = 0; i < copy_count; i++, dst->ptr += dst->step, src += sp->buf->format.bitw)
                            *((uint64_t *)dst->ptr) = *((uint64_t *)src);
                    }
                    break;
                }
            }
        }
        if ((err = soundio_outstream_end_write(outstream)))
        {
            if (err == SoundIoErrorUnderflow)
                return;
            fprintf(stderr, "unrecoverable stream error: %s\n", soundio_strerror(err));
            exit(1);
        }
        if (frames_left <= 0)
            break;
    }

    soundio_outstream_pause(outstream, sp->pause);
}

int cs_speaker_setFormat(CS_Speaker *sp, CS_Format f)
{
    int device_index = _cs_speaker_index_by_id(sp, sp->device_id);
    if (device_index < 0)
        return 1;
    struct SoundIoDevice *device = soundio_get_output_device(sp->io, device_index);
    if (!device)
    {
        return 1;
    }
    if (device->probe_error)
    {
        return 1;
    }

    if (!soundio_device_supports_sample_rate(device, f.srate))
    {
        return 1;
    }
    if (!soundio_device_supports_format(device, _cs_speaker_format(f.bit)))
    {
        return 1;
    }
    if (!soundio_device_supports_layout(device, _cs_speaker_layout(f.layout)))
    {
        return 1;
    }

    sp->fmt = f;
    ele_mixer_setOutputFormat(sp->mixer, f);
}

int cs_speaker_init(CS_Speaker *sp, CS_Format f)
{
    int err = (sp->backend == SoundIoBackendNone) ? soundio_connect(sp->io) : soundio_connect_backend(sp->io, sp->backend);
    if (err)
        return 1;
    soundio_flush_events(sp->io);

    if (0 != cs_speaker_setFormat(sp, f))
    {
        return 1;
    }

    int device_index = _cs_speaker_index_by_id(sp, sp->device_id);
    if (device_index < 0)
        return 1;
    sp->device = soundio_get_output_device(sp->io, device_index);
    if (!sp->device)
    {
        return 1;
    }
    if (sp->device->probe_error)
    {
        return 1;
    }

    sp->outstream = soundio_outstream_create(sp->device);
    if (!sp->outstream)
    {
        return 1;
    }

    sp->outstream->userdata = sp;
    sp->outstream->write_callback = write_callback;
    sp->outstream->underflow_callback = NULL;
    sp->outstream->name = sp->name;
    sp->outstream->software_latency = 0.0f;
    sp->outstream->sample_rate = sp->fmt.srate;

    sp->buf = cs_create_samples(sp->fmt);

    if ((err = soundio_outstream_open(sp->outstream)))
    {
        return 1;
    }

    if (sp->outstream->layout_error)
        return 1;

    return 0;
}

int cs_speaker_play(CS_Speaker *sp)
{
    int err;

    if (!sp || !sp->outstream)
        return 1;

    if ((err = soundio_outstream_start(sp->outstream)))
    {
        fprintf(stderr, "unable to start device: %s\n", soundio_strerror(err));
        return 1;
    }

    for (;;)
        soundio_wait_events(sp->io);
}

int cs_speaker_pause(CS_Speaker *sp, int pause)
{
    if (!sp)
        return 1;

    sp->pause = pause;
    return 0;
}

int cs_speaker_stop(CS_Speaker *s)
{
    if (!s)
        return 1;

    if (s->outstream)
    {
        soundio_outstream_destroy(s->outstream);
        s->outstream = NULL;
    }
    if (s->device)
    {
        soundio_device_unref(s->device);
        s->device = NULL;
    }
    return 0;
}

CS_Speaker *cs_create_speaker()
{
    CS_Speaker *sp = (CS_Speaker *)av_mallocz(sizeof(CS_Speaker));
    if (!sp)
        return NULL;

    sp->io = soundio_create();
    if (!sp->io)
        goto _exit_;

_exit_:
    cs_speaker_destory(&sp);
    return NULL;
}

void cs_speaker_destory(CS_Speaker **sp)
{
    if (!sp || !*sp)
        return;

    CS_Speaker *s = *sp;

    cs_speaker_stop(s);

    if (s->io)
    {
        soundio_destroy(s->io);
        s->io = NULL;
    }
    av_freep(sp);
}