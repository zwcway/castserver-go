
#include "./samples.h"
//#include "./pipeline.h"

typedef struct ELE_Volume
{
    char name[16];

    int power;
    double base;
    double gain;
    double volume;
    int mute;
} ELE_Volume;


ELE_Volume *ele_create_volume(double vol);