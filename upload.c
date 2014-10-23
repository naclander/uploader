#define _GNU_SOURCE
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <sys/stat.h>
#include <stdbool.h>

#include <onion/onion.h>
#include <onion/shortcuts.h>

//Safer asprintf macro from O'Reily (Thanks!)
//Argument must be a char * initialized to NULL
#define Sasprintf(write_to, ...) { \
	char *tmp_string_for_extend = (write_to); \
	asprintf(&(write_to), __VA_ARGS__); \
	free(tmp_string_for_extend); \
}

char * FILE_DIRECTORY = "./files";
char * TEXT_FILE = "text.txt";
char * HTML_PAGE = "index.html";

#define TEXT_LIST_SIZE  10
char * TEXT_LIST[TEXT_LIST_SIZE];

//Attempts to read a specified file
//input:
//file_name: full path to file
//input_file_size: this value will be assigned the size of the file read
//Returns a pointer to the string with the file contents, or null if file not found
char * read_file(char * file_name, unsigned int * input_file_size){
	char * file_contents;
	FILE *input_file = fopen(file_name, "rb");
	if(input_file == NULL){
		return(NULL);	
	}
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

char * edit_html_template(char * current_html_file, char * s1, char * s2,
 					      unsigned int * input_file_size){
	char * new_html_file = NULL;
	Sasprintf(new_html_file,current_html_file,s1,s2);
	*input_file_size = strlen(new_html_file);
	free(current_html_file);
	return(new_html_file);
}

//Parses the text file TEXT_FILE for strings, creates a new formatted string,
//and places the new formatted string in the first %s of current_html_file. It
//wil also set input_file_size to the size of the newly expanded html file.
//current_html_file will be freed
//Input:
//current_html_file: the html file that will have the new formatted text_list string
//placed into it. It must contain two "%s".
//input_file_size: This value will be set to the newly expanded size of the html
//file
//Returns:
//a pointer to the expanded html file. This pointer must be freed. current_html_file
//is freed by this function
char * text_list_to_html(char * current_html_file, unsigned int * input_file_size){
	char * text_list = read_file(TEXT_FILE, input_file_size);
	char * new_html_file = NULL;

	if(text_list == NULL){
		return(edit_html_template(current_html_file,"","%s",input_file_size));
	}

	char * formatted_text_list = NULL;
	Sasprintf(formatted_text_list,"<ul style=\"list-style-type:disc\">");
	char *tok = text_list;
	while((tok = strtok(tok, "\n")) != NULL){
		Sasprintf(formatted_text_list, "%s\n <li> %s </li>",formatted_text_list,tok);
        tok = NULL;
	}
	Sasprintf(formatted_text_list,"%s\n</ul>",formatted_text_list);

	return(edit_html_template(current_html_file,formatted_text_list,"%s",
						      input_file_size));
}

void write_and_free(onion_response * res, char * file, unsigned int input_size){
	onion_response_write(res, file, input_size);
	free(file);
}

onion_connection_status main_page(void *_, onion_request *req, onion_response *res){

	unsigned int input_file_size;
	char * html_file = read_file(HTML_PAGE, &input_file_size);

	html_file = text_list_to_html(html_file, &input_file_size);

	write_and_free(res,html_file,input_file_size*sizeof(char));
	//char * text_list_html = text_list_to_html(&input_file_size);
	//if(text_list_html != NULL){
	//	onion_response_write(res, text_list_html,input_file_size*sizeof(char));
	//}
	//char * files_list = file_list_to_html(&input_file_size);
	//onion_response_write(res, text_list_html,input_file_size*sizeof(char));
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
