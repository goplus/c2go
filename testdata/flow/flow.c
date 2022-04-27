#include <stdio.h>

void f(int a) {
    if (a > 0) {
        a++;
        goto next;
    } else if (a) {
        int b = -3;
        int c;
        (void)c;
        a = b;
        goto next;
    } else {
next:
        printf("Next: %d\n", a);
    }
}

void g(int n) {
    goto one;
    {
        do {
one:        printf("one shot\n");
            break;
        } while (1);
    }
    {
        goto two;
        do {
two:        printf("multiple shots\n");
            n--;
        } while (n > 0);
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
    f(-1);
    g(2);
    return 0;
}
