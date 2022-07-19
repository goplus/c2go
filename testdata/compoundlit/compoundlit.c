#include <stdio.h>

void f(unsigned short a[]) {
    printf("%d, %d, %d\n", a[0], a[1], a[2]);
}

int main() {
    f((unsigned short [3]){ 0x330e, 0, 16 });
    return 0;
}
