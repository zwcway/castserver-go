#ifndef CS_AUDIO
#define CS_AUDIO

#include <stdint.h>
#include <libavutil/samplefmt.h>
#include <libavutil/channel_layout.h>

typedef enum CS_Rate
{
    Rate_NONE = 0,
    Rate_44100,
    Rate_48000,
    Rate_88200,
    Rate_96000,
    Rate_192000,
    Rate_352800,
    Rate_384000,
    Rate_2822400,
    Rate_5644800,
} CS_Rate;

typedef int32_t CS_RateMask;

#define CHANNEL_MAX 32

typedef enum CS_Channel
{
    Channel_NONE = 0,
    Channel_FRONT_LEFT,
    Channel_FRONT_RIGHT,
    Channel_FRONT_CENTER,
    Channel_FRONT_LEFT_OF_CENTER,
    Channel_FRONT_RIGHT_OF_CENTER,
    Channel_LOW_FREQUENCY,
    Channel_BACK_LEFT,
    Channel_BACK_RIGHT,
    Channel_BACK_CENTER,
    Channel_SIDE_LEFT,
    Channel_SIDE_RIGHT,
    Channel_TOP_CENTER,
    Channel_TOP_FRONT_LEFT,
    Channel_TOP_FRONT_CENTER,
    Channel_TOP_FRONT_RIGHT,
    Channel_TOP_BACK_LEFT,
    Channel_TOP_BACK_CENTER,
    Channel_TOP_BACK_RIGHT,
} CS_Channel;

#define _cs_layout_build(l) (1 << (l))
#define _cs_layout_build2(a, b) ((1 << (a)) | (1 << (b)))
#define _cs_layout_extend1(l, a) ((l) | (1 << (a)))
#define _cs_layout_extend2(l, a, b) ((l) | (1 << (a)) | (1 << (b)))

typedef uint32_t CS_ChannelMask;

typedef enum CS_LayoutMask
{
    LayoutMask_10 = _cs_layout_build(Channel_FRONT_CENTER),
    LayoutMask_20 = _cs_layout_build2(Channel_FRONT_LEFT, Channel_FRONT_RIGHT),
#define LayoutMask_Mono LayoutMask_10
#define LayoutMask_Stereo LayoutMask_20
    LayoutMask_21 = _cs_layout_extend1(LayoutMask_20, Channel_LOW_FREQUENCY),
    LayoutMask_22 = _cs_layout_extend2(LayoutMask_20, Channel_SIDE_LEFT, Channel_SIDE_RIGHT),
    LayoutMask_30 = _cs_layout_extend1(LayoutMask_20, Channel_FRONT_CENTER),
    LayoutMask_31 = _cs_layout_extend1(LayoutMask_30, Channel_LOW_FREQUENCY),
    LayoutMask_40 = _cs_layout_extend1(LayoutMask_30, Channel_BACK_CENTER),
    LayoutMask_41 = _cs_layout_extend1(LayoutMask_40, Channel_LOW_FREQUENCY),
    LayoutMask_50 = _cs_layout_extend2(LayoutMask_30, Channel_SIDE_LEFT, Channel_SIDE_RIGHT),
    LayoutMask_51 = _cs_layout_extend1(LayoutMask_50, Channel_LOW_FREQUENCY),
    LayoutMask_5B0 = _cs_layout_extend2(LayoutMask_30, Channel_BACK_LEFT, Channel_BACK_RIGHT),
    LayoutMask_5B1 = _cs_layout_extend1(LayoutMask_5B0, Channel_LOW_FREQUENCY),
    LayoutMask_60 = _cs_layout_extend1(LayoutMask_50, Channel_BACK_CENTER),
    LayoutMask_61 = _cs_layout_extend1(LayoutMask_51, Channel_BACK_CENTER),
    LayoutMask_70 = _cs_layout_extend2(LayoutMask_50, Channel_BACK_LEFT, Channel_BACK_RIGHT),
    LayoutMask_71 = _cs_layout_extend1(LayoutMask_70, Channel_LOW_FREQUENCY),
    LayoutMask_702 = _cs_layout_extend2(LayoutMask_70, Channel_TOP_FRONT_LEFT, Channel_TOP_FRONT_RIGHT),
    LayoutMask_712 = _cs_layout_extend2(LayoutMask_71, Channel_TOP_FRONT_LEFT, Channel_TOP_FRONT_RIGHT),
    LayoutMask_714 = _cs_layout_extend2(LayoutMask_712, Channel_TOP_BACK_LEFT, Channel_TOP_BACK_RIGHT),

} CS_LayoutMask;

typedef enum CS_Bits
{
    Bits_NONE = 0,

    Bits_S8 = 1, // int8
    Bits_U8,     // uint8

    Bits_S16LE = 8, // int16
    Bits_U16LE,     // uint16
    Bits_16LEF,     // float16

    Bits_S24LE = 16, // int24
    Bits_U24LE,      // uint24
    Bits_24LEF,      // float24

    Bits_S32LE = 24, // int32
    Bits_U32LE,      // uint32
    Bits_32LEF,      // float32

    Bits_S64LE = 32, // int64
    Bits_U64LE,      // uint64
    Bits_64LEF,      // float64
} CS_Bits;

#define cs_bits_size(b) (((b) >> 3) + 1)

typedef struct CS_Format
{
    CS_Rate rate : 16;         // 采样率
    uint32_t srate : 32;       // 采样率大小
    CS_Bits bit : 16;          // 样本格式
    uint16_t bitw : 16;        // 单样本字节数
    uint16_t chs : 16;         // 声道数量
    CS_LayoutMask layout : 32; // 声道布局
} CS_Format;

/**
 * @brief 单样本共占用的空间
 *
 */
#define cs_format_size(f, n) ((n) * ((f).bitw) * (f).chs)

#define cs_format_equal(a, b) ((a).rate == (b).rate && (a).bit == (b).bit && (a).layout == (b).layout)

inline void cs_format_merge(CS_Format *dst, CS_Format src)
{
    if (!dst)
        return;

    if (dst->srate < src.srate)
    {
        dst->srate = src.srate;
    }
    if (dst->bit < src.bit)
    {
        dst->bit = src.bit;
        dst->bitw = src.bitw;
    }

    // TODO 取并集
    if (dst->chs < src.chs)
    {
        dst->chs = src.chs;
        dst->layout = src.layout;
    }
}

/**
 * @brief 从内部格式转换到 ffmpeg 支持的格式
 *
 * @param bits
 * @return enum AVSampleFormat
 */
inline enum AVSampleFormat cs_format_bits_to_fmt(CS_Bits bits)
{
    switch (bits)
    {
    case Bits_U8:
        return AV_SAMPLE_FMT_U8P;
    case Bits_S16LE:
        return AV_SAMPLE_FMT_S16P;
    case Bits_S32LE:
        return AV_SAMPLE_FMT_S32P;
    case Bits_32LEF:
        return AV_SAMPLE_FMT_FLTP;
    case Bits_64LEF:
        return AV_SAMPLE_FMT_DBLP;
    default:
        // ffmpeg 不支持的格式先转成 double
        return AV_SAMPLE_FMT_DBLP;
    }
}

/**
 * @brief 从 ffmpeg 支持的格式转换到内部格式
 *
 * @param fmt
 * @return CS_Bits
 */
inline CS_Bits cs_format_fmt_to_bits(enum AVSampleFormat fmt)
{
    switch (fmt)
    {
    case AV_SAMPLE_FMT_U8:
    case AV_SAMPLE_FMT_U8P:
        return Bits_U8;
    case AV_SAMPLE_FMT_S16:
    case AV_SAMPLE_FMT_S16P:
        return Bits_S16LE;
    case AV_SAMPLE_FMT_S32:
    case AV_SAMPLE_FMT_S32P:
        return Bits_S32LE;
    case AV_SAMPLE_FMT_FLT:
    case AV_SAMPLE_FMT_FLTP:
        return Bits_32LEF;
    case AV_SAMPLE_FMT_DBL:
    case AV_SAMPLE_FMT_DBLP:
        return Bits_64LEF;
        return Bits_32LEF;
    case AV_SAMPLE_FMT_S64:
    case AV_SAMPLE_FMT_S64P:
        return Bits_S64LE;
    default:
        return Bits_NONE;
    }
}

inline int64_t cs_format_channel_to_avch(CS_Channel c)
{
    switch (c)
    {
    case Channel_FRONT_LEFT:
        return AV_CH_FRONT_LEFT;
    case Channel_FRONT_RIGHT:
        return AV_CH_FRONT_RIGHT;
    case Channel_FRONT_CENTER:
        return AV_CH_FRONT_CENTER;
    case Channel_LOW_FREQUENCY:
        return AV_CH_LOW_FREQUENCY;
    case Channel_BACK_LEFT:
        return AV_CH_BACK_LEFT;
    case Channel_BACK_RIGHT:
        return AV_CH_BACK_RIGHT;
    case Channel_FRONT_LEFT_OF_CENTER:
        return AV_CH_FRONT_LEFT_OF_CENTER;
    case Channel_FRONT_RIGHT_OF_CENTER:
        return AV_CH_FRONT_RIGHT_OF_CENTER;
    case Channel_BACK_CENTER:
        return AV_CH_BACK_CENTER;
    case Channel_SIDE_LEFT:
        return AV_CH_SIDE_LEFT;
    case Channel_SIDE_RIGHT:
        return AV_CH_SIDE_RIGHT;
    case Channel_TOP_CENTER:
        return AV_CH_TOP_CENTER;
    case Channel_TOP_FRONT_LEFT:
        return AV_CH_TOP_FRONT_LEFT;
    case Channel_TOP_FRONT_CENTER:
        return AV_CH_TOP_FRONT_CENTER;
    case Channel_TOP_FRONT_RIGHT:
        return AV_CH_TOP_FRONT_RIGHT;
    case Channel_TOP_BACK_LEFT:
        return AV_CH_TOP_BACK_LEFT;
    case Channel_TOP_BACK_CENTER:
        return AV_CH_TOP_BACK_CENTER;
    case Channel_TOP_BACK_RIGHT:
        return AV_CH_TOP_BACK_RIGHT;
    }
    return 0;
}

inline CS_Channel cs_format_avch_to_channel(int64_t c)
{
    switch (c)
    {
    case AV_CH_FRONT_LEFT:
        return Channel_FRONT_LEFT;
    case AV_CH_FRONT_RIGHT:
        return Channel_FRONT_RIGHT;
    case AV_CH_FRONT_CENTER:
        return Channel_FRONT_CENTER;
    case AV_CH_LOW_FREQUENCY:
        return Channel_LOW_FREQUENCY;
    case AV_CH_BACK_LEFT:
        return Channel_BACK_LEFT;
    case AV_CH_BACK_RIGHT:
        return Channel_BACK_RIGHT;
    case AV_CH_FRONT_LEFT_OF_CENTER:
        return Channel_FRONT_LEFT_OF_CENTER;
    case AV_CH_FRONT_RIGHT_OF_CENTER:
        return Channel_FRONT_RIGHT_OF_CENTER;
    case AV_CH_BACK_CENTER:
        return Channel_BACK_CENTER;
    case AV_CH_SIDE_LEFT:
        return Channel_SIDE_LEFT;
    case AV_CH_SIDE_RIGHT:
        return Channel_SIDE_RIGHT;
    case AV_CH_TOP_CENTER:
        return Channel_TOP_CENTER;
    case AV_CH_TOP_FRONT_LEFT:
        return Channel_TOP_FRONT_LEFT;
    case AV_CH_TOP_FRONT_CENTER:
        return Channel_TOP_FRONT_CENTER;
    case AV_CH_TOP_FRONT_RIGHT:
        return Channel_TOP_FRONT_RIGHT;
    case AV_CH_TOP_BACK_LEFT:
        return Channel_TOP_BACK_LEFT;
    case AV_CH_TOP_BACK_CENTER:
        return Channel_TOP_BACK_CENTER;
    case AV_CH_TOP_BACK_RIGHT:
        return Channel_TOP_BACK_RIGHT;
    case AV_CH_STEREO_LEFT:
    case AV_CH_STEREO_RIGHT:
    case AV_CH_WIDE_LEFT:
    case AV_CH_WIDE_RIGHT:
        break;
    };
    return Channel_NONE;
}

inline int64_t cs_format_mask_to_layout(CS_LayoutMask m)
{
    switch (m)
    {
    case LayoutMask_10:
        return AV_CH_LAYOUT_MONO;
    case LayoutMask_20:
        return AV_CH_LAYOUT_STEREO;
    case LayoutMask_21:
        return AV_CH_LAYOUT_2POINT1;
    case LayoutMask_22:
        return AV_CH_LAYOUT_2_2;
    case LayoutMask_30:
        return AV_CH_LAYOUT_SURROUND;
    case LayoutMask_31:
        return AV_CH_LAYOUT_3POINT1;
    case LayoutMask_40:
        return AV_CH_LAYOUT_4POINT0;
    case LayoutMask_41:
        return AV_CH_LAYOUT_4POINT1;
    case LayoutMask_50:
        return AV_CH_LAYOUT_5POINT0;
    case LayoutMask_51:
        return AV_CH_LAYOUT_5POINT1;
    case LayoutMask_5B0:
        return AV_CH_LAYOUT_5POINT0_BACK;
    case LayoutMask_5B1:
        return AV_CH_LAYOUT_5POINT1_BACK;
    case LayoutMask_60:
        return AV_CH_LAYOUT_6POINT0;
    case LayoutMask_61:
        return AV_CH_LAYOUT_6POINT1;
    case LayoutMask_70:
        return AV_CH_LAYOUT_7POINT0;
    case LayoutMask_71:
        return AV_CH_LAYOUT_7POINT1;
    case LayoutMask_702:
        return AV_CH_LAYOUT_7POINT0_FRONT;
    case LayoutMask_712:
        return AV_CH_LAYOUT_7POINT1_WIDE;
    case LayoutMask_714:
        return AV_CH_LAYOUT_7POINT1_WIDE_BACK;
    }
}

#endif // !CS_AUDIO
