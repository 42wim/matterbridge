package proxy

/*
#include <stdio.h>

void goCallback_cgo(char * json) {
	printf("inside goCallback_cgo\n");
	void goCallback(char *);
	goCallback(json);
}
*/
import "C"
