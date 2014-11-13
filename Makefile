all:
	clang -g -o upload upload.c -lonion -pthreads

clean:
	rm upload 
