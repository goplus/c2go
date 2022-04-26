#include <stdio.h>

void f(int a) {
    if (a) {
        a++;
        goto next;
    } else {
next:
        printf("Next: %d\n", a);
    }
}

int main() {
    int a = sizeof(int);
    int *b = &a;
    switch (a) {
    case 4:
        printf("Hi\n");
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
        if (a == 1) {
            goto done;
        }
        if (a == 3) {
            printf("a = 3, continue\n");
            continue;
        }
    }
done:
    printf("a = %d\n", a);
    if (b) {
        printf("&a\n");
    }
    f(0);
    f(1);
    return 0;
}
