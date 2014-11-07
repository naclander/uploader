#define _GNU_SOURCE
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

#include <onion/onion.h>
#include <onion/shortcuts.h>
#include <onion/codecs.h>
#include <onion/mime.h>

//Safer asprintf macro from O'Reily (Thanks!)
//Argument must be a char * initialized to NULL
#define Sasprintf(write_to, ...) { \
	char *tmp_string_for_extend = (write_to); \
	asprintf(&(write_to), __VA_ARGS__); \
	free(tmp_string_for_extend); \
}

char * FILE_DIRECTORY = "./files/";
char * TEXT_FILE = "text.txt";
char * HTML_PAGE = "index.html";

char * stream_to_string(FILE * input_stream){
	char * file_contents;
	fseek(input_stream, 0, SEEK_END);
	unsigned int input_file_size = ftell(input_stream);
	rewind(input_stream);
	file_contents = malloc((input_file_size + 1) * (sizeof(char)));
	input_file_size = fread(file_contents, sizeof(char), input_file_size, input_stream);
	file_contents[input_file_size] = 0;
	return(file_contents);
}

//Attempts to read a specified file
//input:
//file_name: full path to file
//input_file_size: this value will be assigned the size of the file read
//Returns a pointer to the string with the file contents, or null if file not found
char * read_file(char * file_name){
	char * file_contents;
	FILE *input_file = fopen(file_name, "rb");
	if(input_file == NULL){
		return(NULL);	
	}
	char * file_string = stream_to_string(input_file);
	fclose(input_file);
	return(file_string);
}

/*
 * Lists all files in FILE_DIRECTORY into an allocated string, and returns it.
 * Input: none
 * Output: Pointer to newly allocated string, NEEDS TO BE FREED
 */
char * file_list_to_html(){
	char * html_file_list = NULL;
	/* TODO find a way to not allocate a string for this */
	char * command = NULL;
	/* TODO Replace this with C directory API (maybe? Discuss) */	
	Sasprintf(command,"ls -t %s",FILE_DIRECTORY);
	FILE * input_stream = popen(command,"r");
	free(command);

	if(input_stream == NULL){
		Sasprintf(html_file_list,"");
	}
	else{
		char * file_list = stream_to_string(input_stream);
		Sasprintf(html_file_list,"<ul style=\"list-style-type:disc\">");
		char *tok = file_list;
		while((tok = strtok(tok, "\n")) != NULL){
			Sasprintf(html_file_list, "%s\n <li> <a href=\"%s?file\"> %s</a></li>",
					  html_file_list,tok,tok);
			tok = NULL;
		}
		Sasprintf(html_file_list,"%s\n</ul>",html_file_list);
	}
	pclose(input_stream);
	return(html_file_list);
}


/*
Parses the text file TEXT_FILE for text, and creates a formatted html reverse text
list string, which it returns.
Input: none
Output: Pointer pointing to the newly allocated string, NEEDS TO BE FREED
*/
char * text_list_to_html(){
	char * text_list = read_file(TEXT_FILE);
	char * html_text_list = NULL;
	char * sanitized_text= NULL;
	if(text_list == NULL){
		Sasprintf(html_text_list,"");
	}
	else{
		Sasprintf(html_text_list,"\n</ul>");
		char *tok = text_list;
		while((tok = strtok(tok, "\n")) != NULL){
			if( (sanitized_text = onion_html_quote(tok) ) != NULL){
				tok = sanitized_text;
			}
			Sasprintf(html_text_list, "<li> %s </li>\n%s",tok,html_text_list);
			sanitized_text == NULL? : free(sanitized_text);
			tok = NULL;
		}
		Sasprintf(html_text_list,"<ul style=\"list-style-type:disc\">\n%s",html_text_list);
	}
	free(text_list);
	return(html_text_list);
}


onion_connection_status main_page(void *_, onion_request *req, onion_response *res){

	if (onion_request_get_query(req, "file")){
		/* TODO: Find a way to not allocate a string for this */
		char * dir_path = NULL;
		char * buffer = NULL;
		FILE * fp = NULL;

		Sasprintf(dir_path,"%s%s",FILE_DIRECTORY,onion_request_get_query(req,"1"));
		if((fp = fopen(dir_path,"rb")) == NULL){
			return(onion_shortcut_response("<b> 404 File Not Found </b>", 404, req, res));
		}
		free(dir_path);
		fseek(fp, 0, SEEK_END);
		int size = ftell(fp);
		fseek(fp, 0, SEEK_SET);
		buffer = malloc(size);
		/*Will break on files of size 0*/
		if (fread(buffer, 1, size, fp) == size && size != 0){
			onion_response_set_header(res, "Content-Type", onion_mime_get(dir_path));
			onion_response_write(res, buffer, size);
		}
		fclose(fp);
		free(buffer);
	}
	else{
		char * html_file = read_file(HTML_PAGE);
		char * html_text_list = text_list_to_html();
		char * html_file_list = file_list_to_html();

		Sasprintf(html_file,html_file,html_text_list,html_file_list);

		onion_response_write(res, html_file, strlen(html_file));

		free(html_file);
		free(html_text_list);
		free(html_file_list);
	}
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
	onion_url_add(urls, "^(.*)$", main_page);
	onion_set_port(o,port);	
	onion_listen(o);
	return(EXIT_SUCCESS);
}
