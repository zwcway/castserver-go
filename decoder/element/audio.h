#ifndef CS_AUDIO
#define CS_AUDIO

#include <stdint.h>
#include <libavutil/samplefmt.h>

typedef enum CS_Channel
{
    Channel_FRONT_LEFT = 1,
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
    Channel_TOP_FRONT_LEFT,
    Channel_TOP_FRONT_CENTER,
    Channel_TOP_FRONT_RIGHT,
    Channel_TOP_BACK_LEFT,
    Channel_TOP_BACK_CENTER,
    Channel_TOP_BACK_RIGHT,
} CS_Channel;

typedef enum CS_Bits
{
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

    Bits_64LEF = 32, // float64
} CS_Bits;

#define cs_bits_size(b) (((b) >> 3) + 1)

typedef struct CS_Format
{
    uint32_t rate;   // 采样率
    CS_Bits bit : 8; // 样本格式
    uint8_t bitw;    // 单样本字节数
    uint16_t chs;    // 声道数量
    uint64_t layout; // 声道布局
} CS_Format;

/**
 * @brief 单样本共占用的空间
 * 
 */
#define cs_format_size(f, n) ((n) * ((f).bitw) * (f).chs)

inline void cs_format_merge(CS_Format *dst, CS_Format src)
{
    if (!dst)
        return;

    if (dst->rate < src.rate)
    {
        dst->rate = src.rate;
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
        return AV_SAMPLE_FMT_NONE;
    }
}

#endif // !CS_AUDIO
