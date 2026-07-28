#ifndef ELE_PIPELINE
#define ELE_PIPELINE

#include "./samples.h"


/**
 * @brief 流处理器入口
 * 
 */
typedef  int (*Func_Stream)(void *, CS_Samples *);

/**
 * @brief 获取音频格式
 * 
 */
typedef  int (*Func_AudioFormat)(void *, CS_Format *);


typedef struct CS_Streamer
{
    void *ele;
    int *power;
    Func_Stream stream;

    int cost;
} CS_Streamer;

#define ELE_PIPELINE_STREAM_SIZE 15

typedef struct ELE_Pipeline
{
    CS_Samples *buf;

    CS_Streamer *eles[ELE_PIPELINE_STREAM_SIZE + 1];
    int size;

    int cost;
    int maxCost;
} ELE_Pipeline;


ELE_Pipeline *ele_create_pipeline();
void ele_pipeline_destory(ELE_Pipeline **pp);
int ele_pipeline_add(ELE_Pipeline *p, CS_Streamer *s);
int ele_pipeline_stream(ELE_Pipeline *p, CS_Samples *s);



#endif // !ELE_PIPELINE