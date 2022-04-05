#include <stdio.h>

typedef struct {
    char msg[10];
} foo_t;

int main() {
    foo_t foo = {"Hi, c2go!"};
    foo_t *pfoo = &foo;
    char msg[] = {'a', 'b', '\0'};
    char *pmsg = msg;
    printf("%c\n", msg[1]);
    pmsg[1] = '!', printf("%s\n", pmsg),
    pfoo->msg[2] = '!', printf("%s\n", foo.msg);
    return 0;
}
