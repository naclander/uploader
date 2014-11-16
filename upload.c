#define _GNU_SOURCE
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
/* Maybe replace with threads.h when support exists */
#include <pthread.h>

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
size_t MAX_POST_SIZE = 2<<26; /* 134.218 Megabytes */
size_t FILE_TTL = 120; /* How long the server keeps files and strings, in seconds */

size_t get_stream_size(FILE * f){
	fseek(f,0,SEEK_END);
	size_t size = ftell(f);
	rewind(f);
	return(size);
}

char * stream_to_buffer(FILE * input_stream){
	char * file_contents;
	size_t input_file_size = get_stream_size(input_stream);
	/* If its a pipe stream, cap the file size at MAX_POST_SIZE */
	int magic_number = MAX_POST_SIZE;
	if(input_file_size > magic_number){
		input_file_size = magic_number;
	}
	file_contents = malloc((input_file_size + 1));
	input_file_size = fread(file_contents, 1, input_file_size, input_stream);
	file_contents[input_file_size] = 0;
	return(file_contents);
}

//Attempts to read a specified file
//input:
//file_name: full path to file
//input_file_size: this value will be assigned the size of the file read
//Returns a pointer to the string with the file contents, or null if file not found
char * read_file(char * file_name){
	FILE *input_file = fopen(file_name, "rb");
	if(input_file == NULL){
		return(NULL);	
	}
	char * file_string = stream_to_buffer(input_file);
	fclose(input_file);
	return(file_string);
}

char * file_list_to_string(){
	/* TODO Replace this with C directory API (maybe? Discuss) */	
	/* TODO find a way to not allocate a string for this */
	char * command = NULL;
	char * file_list = NULL;
	Sasprintf(command,"ls -t %s",FILE_DIRECTORY);
	FILE * input_stream = popen(command,"r");
	if(input_stream != NULL){
		free(command);
		file_list = stream_to_buffer(input_stream);	
	}
	pclose(input_stream);
	return(file_list);
}

/*
 * Lists all files in FILE_DIRECTORY into an allocated string, and returns it.
 * Input: none
 * Output: Pointer to newly allocated string, NEEDS TO BE FREED
 */
char * file_list_to_html(){
	char * html_file_list = NULL;
	char * file_list = file_list_to_string();	
	if(file_list == NULL){
		/* Empty html_file_list */
		html_file_list[0] = 0;
	}
	else{
		Sasprintf(html_file_list,"<ul style=\"list-style-type:disc\">");
		char *tok = file_list;
		while((tok = strtok(tok, "\n")) != NULL){
			Sasprintf(html_file_list, "%s\n <li> <a href=\"%s?file\"> %s</a></li>",
					  html_file_list,tok,tok);
			tok = NULL;
		}
		Sasprintf(html_file_list,"%s\n</ul>",html_file_list);
		free(file_list);
	}
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
		/* Empty html_file_list */
		html_text_list[0] = 0;
	}
	else{
		Sasprintf(html_text_list,"\n</ul>");
		char *tok = text_list;
		while((tok = strtok(tok, "\n")) != NULL){
			if( (sanitized_text = onion_html_quote(tok) ) != NULL){
				tok = sanitized_text;
			}
			Sasprintf(html_text_list, "<li> %s </li>\n%s",tok,html_text_list);
			free(sanitized_text);
			tok = NULL;
		}
		Sasprintf(html_text_list,"<ul style=\"list-style-type:disc\">\n%s",html_text_list);
	}
	free(text_list);
	return(html_text_list);
}


void handle_file_query(onion_request * req, onion_response * res){
	/* TODO: Find a way to not allocate a string for this */
	char * dir_path = NULL;
	char * buffer = NULL;
	FILE * fp = NULL;

	Sasprintf(dir_path,"%s%s",FILE_DIRECTORY,onion_request_get_query(req,"1"));
	if((fp = fopen(dir_path,"rb")) == NULL){
		/* TODO: Don't hard code 404 */
		onion_shortcut_response("<b> 404 File Not Found </b>", 404, req, res);
		return;
	}
	free(dir_path);
	size_t size = get_stream_size(fp);
	buffer = malloc(size);
	/*TODO: Will break on files of size 0*/
	if (fread(buffer, 1, size, fp) == size && size != 0){
		onion_response_set_header(res, "Content-Type", onion_mime_get(dir_path));
		onion_response_write(res, buffer, size);
	}
	fclose(fp);
	free(buffer);
}

void handle_main_page(onion_request * req, onion_response * res){
	char * html_file = read_file(HTML_PAGE);
	char * html_text_list = text_list_to_html();
	char * html_file_list = file_list_to_html();

	Sasprintf(html_file,html_file,html_text_list,html_file_list);

	onion_response_write(res, html_file, strlen(html_file));

	free(html_file);
	free(html_text_list);
	free(html_file_list);
}

onion_connection_status main_page(void *_, onion_request *req, onion_response *res){

	onion_request_get_query(req,"file") ? handle_file_query(req,res) :
										  handle_main_page(req,res);
	return(OCS_PROCESSED);
}

onion_connection_status post_data(void *_, onion_request *req, onion_response *res){
	const char *user_data=onion_request_get_post(req,"text");
	const char * filename = onion_request_get_post(req,"file");
	char * file_buffer = malloc(MAX_POST_SIZE);

	if(user_data != NULL){
		/* TODO: Find a way to not allocate a string for this */
		char * command = NULL;
		FILE * input_stream = fopen(TEXT_FILE,"a");
		if(fprintf(input_stream,"%s\n",user_data) < 0){
			printf("Couldn't append to file\n");
			exit(EXIT_FAILURE);
		}
		fclose(input_stream);
		free(command);
	}
	if(filename != NULL){
		/* TODO: Verify filename is in correct format? */
		printf("Filename: %s\n",filename);
		char * file_buffer = NULL;
		char * file_path = NULL;
		FILE * new_file = NULL;
		FILE * uploaded_file= fopen(onion_request_get_file(req,"file"),"r");
		if(!uploaded_file){
			printf("Couldn't open new file\n");
			exit(EXIT_FAILURE);
		}
		/* TODO: Get the file path without allocating a new string */
		Sasprintf(file_path,"%s/%s",FILE_DIRECTORY,filename);	
		new_file = fopen(file_path,"w+b");
		size_t file_size = get_stream_size(uploaded_file);
		size_t size_wrote = 0;
		file_buffer = stream_to_buffer(uploaded_file);	
		if((size_wrote = fwrite(file_buffer,1,file_size,new_file)) < file_size){
			printf("Couldn't write new file\n");
			printf("File size: %zu, size wrote: %zu\n",file_size,size_wrote);
			exit(EXIT_FAILURE);
		}
		free(file_path);
		free(file_buffer);
		fclose(uploaded_file);
		fclose(new_file);
	}
	/* Redirect so we can refresh without resending the form */
	return(onion_shortcut_redirect("/",req,res));
}

void delete_files(){

	while(1){
		unsigned int current_time = time(NULL);
		char * file_list = NULL;
		file_list = file_list_to_string();
		char *file = file_list;
		while((file = strtok(file, "\n")) != NULL){
			char * file_path = NULL;
			Sasprintf(file_path,"%s/%s",FILE_DIRECTORY,file);
			char * command = NULL;
			Sasprintf(command,"date -r %s +%%s",file_path);	
			FILE * command_stream = popen(command,"r");
			free(command);
			char * file_time_string = stream_to_buffer(command_stream);
			unsigned int file_time = atoi(file_time_string);
			free(file_time_string);
			pclose(command_stream);
			if( current_time > (file_time + FILE_TTL)){
				/* TODO: Use onion logging for this */
				printf("current time: %u, file_time: %u, FILE_TTL: %u\n",current_time,
						file_time,FILE_TTL);
				unlink(file_path);
			}
			else{
				free(file_path);
				break;
			}
			free(file_path);
			file = NULL;
		}
		sleep(FILE_TTL/2);
	}
}

int main(int argc, char **argv){
	pthread_t deletion_thread;
	pthread_create(&deletion_thread,NULL,(void * ) &delete_files, NULL);
	char * port = "8080";
	onion * server = onion_new(O_ONE_LOOP);
	onion_url * urls = onion_root_url(server);
	onion_set_max_file_size(server,MAX_POST_SIZE);
	/* TODO: Remove magic number 128 */
	onion_set_max_post_size(server,128);
	onion_url_add(urls, "data", post_data);
	onion_url_add(urls, "", main_page);
	onion_url_add(urls, "^(.*)$", main_page);
	onion_set_port(server,port);	
	onion_listen(server);
	return(EXIT_SUCCESS);
}
