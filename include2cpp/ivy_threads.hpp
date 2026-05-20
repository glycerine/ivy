#pragma once

#include <cstdlib>
#include <iostream>

#ifdef _WIN32
#include <windows.h>
#else
#include <pthread.h>
#include <time.h>
#endif

class reader {
public:
    virtual int fdes() = 0;
    virtual void read() = 0;
    virtual void bind() {}
    virtual bool running() { return fdes() >= 0; }
    virtual bool background() { return false; }
    virtual ~reader() {}
};

class timer {
public:
    virtual int ms_delay() = 0;
    virtual void timeout(int) = 0;
    virtual ~timer() {}
};

#ifdef _WIN32
static DWORD WINAPI ReaderThreadFunction(LPVOID lpParam) {
    reader *cr = static_cast<reader *>(lpParam);
    cr->bind();
    while (true) {
        cr->read();
    }
    return 0;
}

static DWORD WINAPI TimerThreadFunction(LPVOID lpParam) {
    timer *cr = static_cast<timer *>(lpParam);
    while (true) {
        int ms = cr->ms_delay();
        Sleep(ms);
        cr->timeout(ms);
    }
    return 0;
}
#else
static void *_thread_reader(void *rdr_void) {
    reader *rdr = static_cast<reader *>(rdr_void);
    rdr->bind();
    while (rdr->running()) {
        rdr->read();
    }
    delete rdr;
    return 0;
}

static void *_thread_timer(void *tmr_void) {
    timer *tmr = static_cast<timer *>(tmr_void);
    while (true) {
        int ms = tmr->ms_delay();
        struct timespec ts;
        ts.tv_sec = ms / 1000;
        ts.tv_nsec = (ms % 1000) * 1000000;
        nanosleep(&ts, 0);
        tmr->timeout(ms);
    }
    return 0;
}
#endif
