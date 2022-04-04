#include <stdio.h>

int main() {
    char msg[] = {'a', 'b', '\0'};
    char *pmsg = msg;
    printf("%c\n", msg[1]);
    pmsg[1] = '!';
    printf("%s\n", pmsg);
    return 0;
}
