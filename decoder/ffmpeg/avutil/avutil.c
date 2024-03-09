#include <libavformat/avformat.h>
#include <libavutil/avutil.h>
#include <libavutil/samplefmt.h>

static int go_averror_is_eof(int code)
{
    return code == AVERROR_EOF;
}
static void *go_malloc(int size)
{
    return av_malloc(size);
}
