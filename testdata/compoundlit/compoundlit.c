#include <stdio.h>

void f(unsigned short a[]) {
    printf("%d, %d, %d\n", (int)a[0], (int)a[1], (int)a[2]);
}

void g(unsigned short (*a)[3]) {
    printf("%d, %d, %d\n", (int)(*a)[0], (int)(*a)[1], (int)(*a)[2]);
}

int main() {
    f((unsigned short [3]){ 0x330e, 0, 16 });
    g(&(unsigned short [3]){ 0x330e, 0, 16 });
    return 0;
}
