#include <stdio.h>
#include <stdlib.h>
#include <math.h>

void a(int, int, int);
void b(int, int, int);
void c(int, int);
void d(int, int);

void a(int x, int y, int z) {
    b(x, y, z);
}

void b(int x, int y, int z) {
    c(x, y);
}

void c(int x, int y) {
	d(x, y);
}

void d(int x, int y) {
	u_int64_t *s1 = malloc(10000 * sizeof(char));
	for (int i = 0; i < 10000; i++) {
		s1[i] = (u_int64_t)i;
	}
	double a = 1.234;
	double b = 5.7865;
	while(1) {
		a = a + b;
		b = b * a;
    }
	free(s1);
}

int main(int argc, char* argv[]) {
	int x = 10;
	int y = 11;
	int z = 50;
	a(x, y, z);
	return 0;
}
