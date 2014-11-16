all:
	gcc -g -o upload upload.c -lonion -lpthread -Wall

clean:
	rm upload 
