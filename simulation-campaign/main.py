import subprocess
import time
import json
import os
import shutil

from statistics import mean, stdev
from math import sqrt
from collections import defaultdict
import matplotlib.pyplot as plt

from matplotlib.ticker import ScalarFormatter

from configuration import *

DOCKER_RUN_PREFIX = "sudo docker run --rm --security-opt=seccomp:unconfined pzierahn/omnetpp_offload " 

COMMAND_BROKER_RUN = "nohup opp_offload_broker > opp_offload_broker.log 2>&1 &"
COMMAND_SCENARIO_RUN = "/usr/local/go/bin/go run {project_dir}/cmd/evaluation/scenario.go -scenario {sim_name} -broker {broker} -repeat {repetition} -connect {conn_type} -worker {worker_config} -simulation {project_dir}/evaluation/{app_name}"

COMMAND_PROVIDER_RUN = "/home/doganalp/go/bin/opp_offload_worker -broker %s -jobs %d -name native-%d 2>&1 &"
COMMAND_PROVIDER_RUN_DOCKER = DOCKER_RUN_PREFIX + COMMAND_PROVIDER_RUN

def create_chart(x, xlb, y, ylb, path, tit="", err=1, markers=MARKERS, colors=COLORS):
	plt.rcParams['axes.labelsize'] = 24
	plt.rcParams['xtick.labelsize'] = 18
	plt.rcParams['ytick.labelsize'] = 18
	plt.tick_params(axis='both', which='minor', labelsize=10)
	plt.clf()
	fig, ax = plt.subplots()
	ax.spines["right"].set_visible(False)
	ax.spines["top"].set_visible(False)

	for i in range(len(y)):
		if not err:
			plt.plot(x, y[i][0], label=y[i][1], markersize=12, lw=1.2)
		else:
			plt.errorbar(x, y[i][0], yerr=y[i][2], marker=markers[i], label=y[i][1], markersize=12, lw=1.2)

	plt.xticks(x)
	plt.xlabel(xlb)
	plt.ylabel(ylb)
	plt.title(tit)
	plt.xlim(x[0] - x[0] * 0.05, x[-1] + x[0] * 0.05)
	#plt.ylim(0, 60, 5)

	if len(y) > 1:
		plt.legend(loc='best', prop={'size': 20}, framealpha=0.5)
	
	plt.savefig(path + ".png", bbox_inches='tight')
	plt.savefig(path + ".eps", bbox_inches='tight')

def evaluate_conf_int(lst):
	return 1.96*stdev(lst)/sqrt(len(lst))

def move_results():
	for conn_type in CONNECTION_TYPE_LIST:		
		for is_docker in [True, False]:
			for instance in PARALLEL_INSTANCE_LIST:
				simulation_name = APPLICATION_NAME + "-d_%s,j_%d,c_%s" %(is_docker, instance, conn_type)
				result_file = GO_EVALUATION_DIR + simulation_name + "/scenario_documentation.json"
				if os.path.isfile(result_file):
					shutil.copy(result_file, EVALUATION_DIR + "/" + simulation_name + ".json")

def process_result(sim):
	result_file_name = EVALUATION_DIR + str(sim) + ".json"
	results = json.load(open(result_file_name, 'r'))

	if len(results["trails"]) > 1:
		results = list(map(float, [result["end"] - result["start"] for result in results["trails"]]))
		return mean(results), evaluate_conf_int(results)
	else:
		return 0, 0

def draw_eval_local(is_omnet=True):

	move_results()

	omnet_results = []

	if is_omnet:
		mean_list, conf_int_list = [], []
		for instance in PARALLEL_INSTANCE_LIST:
			simulation_name = APPLICATION_NAME + "-d_none,j_%d,c_none" %(instance)
			mean_val, conf_int = process_result(simulation_name)
			mean_list.append(mean_val / 1000.0)
			conf_int_list.append(conf_int / 1000.0)
		omnet_results.append([mean_list, "standalone", conf_int_list]) 	

	for conn_type in CONNECTION_TYPE_LIST:		
		results = [r for r in omnet_results]

		for is_docker in [True, False]:
			mean_list, conf_int_list = [], []
			for instance in PARALLEL_INSTANCE_LIST:
				simulation_name = APPLICATION_NAME + "-d_%s,j_%d,c_%s" %(is_docker, instance, conn_type)
				
				mean_val, conf_int = process_result(simulation_name)
				mean_list.append(mean_val / 1000.0)
				conf_int_list.append(conf_int / 1000.0)		

			results.append([mean_list, "docker" if is_docker else "native", conf_int_list]) 

		create_chart(PARALLEL_INSTANCE_LIST, "Number of parallel instances", results, "Completion time (s)", "./figures/local_" + str(conn_type))

def create_provider_profile(count, cores, is_docker):

	with open("workers.json", "w") as config:

		configs = []

		name = "docker" if is_docker else "native"

		for i in range(count):
			worker = {"jobs": cores, "docker": is_docker, "name": name + "-%d" %(i), "broker": BROKER}
			configs.append(worker)

		json.dump(configs, config)

def evaluate_omnet_instance(instance, project_dir=PROJECT_DIR, project_name=APPLICATION_NAME, project_config="TicToc18"):

	project_dir = project_dir + "/evaluation/" + project_name	
	os.chdir(project_dir)

	sim_name = "tictoc-d_none,j_%d,c_none" %(instance)

	result_file_name = EVALUATION_DIR + sim_name + '.json'
	result_file = open(result_file_name, 'w+', newline='')

	result = {"trails": [], "repeat": NUM_REPETITION, "name": sim_name}

	for i in range(NUM_REPETITION):
		starting_time = time.time() * 1000
		#clean
		output = subprocess.check_output("make cleanall", shell=True)
		print(output)
		#clean
		output = subprocess.check_output("opp_makemake -f --deep -u Cmdenv -o %s" %(project_name), shell=True)
		print(output)
		# compile
		output = subprocess.check_output("make -j %d MODE=release" %(instance), shell=True)
		print(output)

		print("Compilation is done...")
		print("Running repetition %d with instances %d" %(i, instance))
		# run
		output = subprocess.check_output("opp_runall -j %d ./%s -c %s" %(instance, project_name, project_config), shell=True)
		end_time = time.time() * 1000
		result["trails"].append({"index": i, "start": starting_time, "end": end_time})
		
	json.dump(result, result_file)
	result_file.close()

def evaluate_omnet(project_dir=PROJECT_DIR, project_name=APPLICATION_NAME, project_config="TicToc18"):

	for instance in PARALLEL_INSTANCE_LIST:
		evaluate_omnet_instance(instance, project_dir=project_dir, project_name=project_name, project_config=project_config)

def evaluate_local(is_omnet=False):

	for instance in PARALLEL_INSTANCE_LIST:
		if is_omnet:
			evaluate_omnet_instance(instance)
		for conn_type in CONNECTION_TYPE_LIST:
			for is_docker in [False]:
				simulation_name = APPLICATION_NAME + "-d_%s,j_%d,c_%s" %(is_docker, instance, conn_type)

				create_provider_profile(NUM_PROVIDER, instance, is_docker)

				command = COMMAND_SCENARIO_RUN.format(project_dir=PROJECT_DIR, sim_name=simulation_name, 
					broker=BROKER, repetition=NUM_REPETITION, 
					conn_type=conn_type, worker_config=PROJECT_DIR + "simulation-campaign/workers.json", 
					app_name=APPLICATION_NAME)
				subprocess.check_output(command, shell=True)

#move_results()
evaluate_local(is_omnet=False)
#draw_eval_local()
#evaluate_omnet()
