#include <stdio.h>

typedef struct {
    unsigned short a :1,
        b :2,
          :3,
        c :5;
    short x :1,
        y :2,
          :3,
        z :5;
} foo;

typedef long time_t;

typedef struct {
    time_t tv_sec;
    int :8*(sizeof(time_t)-sizeof(long))*(1234==4321);
    long tv_nsec;
    int :8*(sizeof(time_t)-sizeof(long))*(1234!=4321);
} timespec_t;

int main() {
    foo foo;
    foo.a = 1;
    foo.b = 3;
    foo.c = 15;
    foo.z = 15;
    foo.x = 1;
    foo.y = -1;
    printf(
        "foo.a = %d, foo.b = %d, foo.c = %d\n",
        foo.a, foo.b, foo.c);
    printf(
        "foo.x = %d, foo.y = %d, foo.z = %d\n",
        foo.x, foo.y, foo.z);
    return 0;
}
