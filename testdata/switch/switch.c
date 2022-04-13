#include <stdio.h>

int main() {
    int c = 9;
    switch (c&3) while((c-=4)>=0) {
        printf("=> %d\n", c);
        case 3: printf("3\n");
        case 2: printf("2\n");
            break;
        default: printf("default\n");
        case 0: printf("0\n");
    }
    return 0;
}
