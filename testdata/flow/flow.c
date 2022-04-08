#include <stdio.h>

int main() {
    int a = sizeof(1);
    int *b = &a;
    switch (a) {
    case 4:
        printf("sizeof(int) == 4\n");
        break;
    case 8:
        printf("sizeof(int) == 8\n");
        break;
    case 0:
    case 1:
    default:
        printf("sizeof(int) == unknown\n");
    }
    while (a) {
        a--;
        if (a == 3) {
            goto done;
        }
        if (a == 5) {
            printf("a = 5, continue\n");
            continue;
        }
    }
done:
    printf("a = %d\n", a);
    if (b) {
        printf("&a\n");
    }
    return 0;
}
