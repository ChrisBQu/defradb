cmake_minimum_required(VERSION 3.10)
project(nativewrapper)

# Don't hardcode NDK path - let Gradle handle it
# Remove these lines:
# set(ANDROID_NDK_HOME "/home/revlinux/Android/Sdk/ndk/29.0.13113456")
# set(CMAKE_TOOLCHAIN_FILE ${ANDROID_NDK_HOME}/build/cmake/android.toolchain.cmake)

# Let Gradle pass these variables
# set(ANDROID_ABI "arm64-v8a")  # Gradle will pass this
# set(ANDROID_PLATFORM "android-29")  # Gradle will pass this

set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -std=c++11")

# Build your wrapper
add_library(nativewrapper SHARED
    src/main/cpp/jni_bridge.c
)

# Include headers
include_directories(${CMAKE_SOURCE_DIR}/src/main/cpp)

# Import prebuilt library
add_library(core SHARED IMPORTED)
set_target_properties(core PROPERTIES
    IMPORTED_LOCATION ${CMAKE_SOURCE_DIR}/src/main/jniLibs/arm64-v8a/libdefradb.so
)

target_link_libraries(nativewrapper
    core
    log
)