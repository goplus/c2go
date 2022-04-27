#include <stdio.h>

int f(int n) {
    return n;
}

void call(void (*f)()) {
    if (f == (void(*)())-1) {
        return;
    }
    printf("n: %d\n", ((int(*)(int))f)(3));
}

int main() {
    call((void(*)())-1);
    call((void*)f);
    return 0;
}
