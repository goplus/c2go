#include <stdio.h>

int main() {
    long long a = 3;
    __atomic_store_n(&a, 100, 0);
    printf("atomic: %lld\n", a);
    return 0;
}