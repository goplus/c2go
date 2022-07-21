#include <stdio.h>

#define aslong(v) ((union{double _fval; long _ival;}){v})._ival

void test() {
    if (aslong(1.1)) {
        printf("%d\n", (int)aslong(1.1));
    }
}

void f(unsigned short a[]) {
    printf("%d, %d, %d\n", (int)a[0], (int)a[1], (int)a[2]);
}

void g(unsigned short (*a)[3]) {
    printf("%d, %d, %d\n", (int)(*a)[0], (int)(*a)[1], (int)(*a)[2]);
}

int main() {
    f((unsigned short [3]){ 0x330e, 0, 16 });
    g(&(unsigned short [3]){ 0x330e, 0, 16 });
    test();
    return 0;
}
