#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <sys/stat.h>
#include <stdbool.h>

#include <onion/onion.h>
#include <onion/shortcuts.h>

char * FILE_DIRECTORY = "./files";
char * HTML_PAGE = "index.html";

#define TEXT_LIST_SIZE  10
char * TEXT_LIST[TEXT_LIST_SIZE];

char * read_file(char * file_name, unsigned int * input_file_size){
	char * file_contents;
	FILE *input_file = fopen(file_name, "rb");
	fseek(input_file, 0, SEEK_END);
	*input_file_size = ftell(input_file);
	rewind(input_file);
	file_contents = malloc((*input_file_size + 1) * (sizeof(char)));
	fread(file_contents, sizeof(char), *input_file_size, input_file);
	fclose(input_file);
	file_contents[*input_file_size] = 0;
	return(file_contents);
}

onion_connection_status post_data(void *_, onion_request *req, onion_response *res){
	if (onion_request_get_flags(req)&OR_HEAD){
		onion_response_write_headers(res);
		return OCS_PROCESSED;
	}
	//user_data could be null if "text" form was not filled
	const char *user_data=onion_request_get_post(req,"text");
	if(!strcmp(user_data,"")){
		onion_response_printf(res, "No User data");
	}
	else{
		onion_response_printf(res, "The user wrote: %s", user_data);
	}
	return OCS_PROCESSED;
}

bool file_exists(char * file_path){
	struct stat s;
	if(-1 == stat(file_path, &s)){
		return(false);
	}
	return(true);
}

onion_connection_status main_page(void *_, onion_request *req, onion_response *res){
	unsigned int input_file_size;
	char * html_file = read_html_file("index.html", &input_file_size);
	onion_response_write(res, html_file,input_file_size*sizeof(char));
	return OCS_PROCESSED;
}

int main(int argc, char **argv){
	char * port = "8080";
	onion * o=onion_new(O_ONE_LOOP);
	onion_url * urls=onion_root_url(o);
	
	/*
	onion_url_add_static(urls, "", 
"<html>\n"
"<head>\n"
" <title>Simple post example</title>\n"
"</head>\n"
"\n"
"Write something: \n"
"<form method=\"POST\" action=\"data\">\n"
"<input type=\"text\" name=\"text\">\n"
"<input type=\"submit\">\n"
"</form>\n"
"\n"
"</html>\n", HTTP_OK);
	*/

	onion_url_add(urls, "data", post_data);
	onion_url_add(urls, "", main_page);
	onion_set_port(o,port);	
	onion_listen(o);
}
