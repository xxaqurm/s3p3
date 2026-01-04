#include <iostream>
#include <sstream>
#include <fstream>

#include <ctime>

#include "jsonParser.hpp"
#include "jsonValue.hpp"
#include "queryParser.hpp"

#include "hashMap.hpp"
#include "database.hpp"

using namespace std;

int main(int argc, char** argv) {
	ofstream logfile("log.txt", ios::app);
	
	time_t current_time = time(nullptr);
	logfile << "Time: " << asctime(localtime(&current_time));
	

	DBCommand cmd = parseQuery(argc, argv);

	logfile << "Request received: "
		 	<< cmd.database << " "
			<< cmd.collection << " "
		 	<< "hope its my action" << " "
		 	<< cmd.json << "\n\n";

	JSONNode document = loadCollection(cmd.database, cmd.collection);
	executeCommand(cmd, document);
	
	if (cmd.action != CommandAction::FIND) {
		saveCollection(cmd.database, cmd.collection, document);
	}
}
