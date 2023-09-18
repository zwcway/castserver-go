cmake_minimum_required(VERSION 3.7)

project(speaker_test VERSION 0.1 LANGUAGES C)

find_package(PkgConfig REQUIRED)
pkg_check_modules(AVCODEC REQUIRED IMPORTED_TARGET libavcodec)
pkg_check_modules(AVFORMAT REQUIRED IMPORTED_TARGET libavformat)
#pkg_check_modules(AVFILTER		REQUIRED IMPORTED_TARGET libavfilter)
#pkg_check_modules(AVDEVICE		REQUIRED IMPORTED_TARGET libavdevice)
pkg_check_modules(AVUTIL REQUIRED IMPORTED_TARGET libavutil)
pkg_check_modules(SWRESAMPLE		REQUIRED IMPORTED_TARGET libswresample)
#pkg_check_modules(SWSCALE		REQUIRED IMPORTED_TARGET libswscale)

find_package(Soundio REQUIRED)

add_executable(${PROJECT_NAME} speaker_test.c)
target_link_libraries(${PROJECT_NAME} 
    PkgConfig::AVCODEC
    PkgConfig::AVFORMAT
    #    PkgConfig::AVFILTER
    #    PkgConfig::AVDEVICE
    PkgConfig::AVUTIL
    PkgConfig::SWRESAMPLE
    #    PkgConfig::SWSCALE
)
