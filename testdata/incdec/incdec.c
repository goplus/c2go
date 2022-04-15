#include <stdio.h>

int main() {
    int a = 1;
    int *pa = &a;
    (*(0+pa))++;
    printf("%d\n", a);
    return 0;
}
