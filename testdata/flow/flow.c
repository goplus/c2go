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

void h() {
    for (int i = 0; i < 3; i++) {
        printf("Hi %d\n", i);
    }
    int i = 1;
    for (int j; i < 3; i++) {
        (void)j;
        printf("Hi %d\n", i);
    }
}

struct foo {
    int a;
    double b;
};

int main() {
    int a = __builtin_offsetof(struct foo, b);
    int *b = &a;
    switch (a) {
    case 4:
        printf("Hi\n");
        printf("offsetof(b) == 4\n");
        break;
    case 8:
        printf("offsetof(b) == 8\n");
        break;
    case 0:
    case 1:
    default:
        printf("offsetof(b) == unknown\n");
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
    h();
    return 0;
}
