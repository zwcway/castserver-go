#ifndef CS_SAMPLES
#define CS_SAMPLES

#include <libavutil/avutil.h>
#include "audio.h"

/**
 * @brief 预留 32 个声道
 *
 */
#define BUFFER_OFFSET 256

#define SAMPLES_BY_DURATION(r, d) ((r) * (d) / 1000)

int config_buffer_duration;

typedef struct CS_Samples
{
    int req_nb_samples;  // 当前样本数量
    int last_nb_samples; // 最近一次处理后的样本数量

    CS_Format format; // 当前格式

    CS_Format real_fmt; // 初始格式
    int real_size;      // 初始字节数

    double **data;
    uint8_t **raw_data;
    uint8_t *ptr[CHANNEL_MAX]; // 临时保存正在处理中的样本数组

    int auto_size;

    uint8_t channel_index[CHANNEL_MAX];
} CS_Samples;

#define cs_samples_slience(s) memset((s)->raw_data, 0, (s)->real_size)

/**
 * @brief 创建样本缓存
 *
 * @param duration
 * @param rate
 * @param bit
 * @param chs
 * @return CS_Samples*
 */
CS_Samples *cs_create_samples(CS_Format f);
void cs_samples_destory(CS_Samples **s);

/**
 * @brief 复制
 *
 * @param s
 * @param src
 * @param size
 * @param f
 * @return int
 */
int cs_samples_copy_from(CS_Samples *s, const uint8_t *const src, int nb_samples, const CS_Format f);

/**
 * @brief
 *
 * @param s
 * @param dst
 * @param ch
 * @param index
 */
void cs_samples_copy_step(CS_Samples *s, void *dst, int ch, int index);

int cs_samples_default_size(CS_Format f);

#endif // !CS_SAMPLES