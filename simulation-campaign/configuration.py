
PROJECT_DIR = "/home/doganalp/Workspace/omnetpp_offload/"
EVALUATION_DIR = "/home/doganalp/Workspace/omnetpp_offload/simulation-campaign/evaluation/"

GO_EVALUATION_DIR = "/home/doganalp/.cache/omnetpp-offload/evaluation/"

# does not affect if it is not local
IS_DOCKER = True

APPLICATION_NAME = "tictoc"

NUM_REPETITION = 7
NUM_PROVIDER = 1
JOBS = 2 # how many cores per provider

CONNECTION_TYPE = "relay" # relay

BROKER = "--"
#BROKER = "192.168.0.233"


PARALLEL_INSTANCE_LIST = [1, 2, 3, 4, 5]
CONNECTION_TYPE_LIST = ["relay"]

MARKERS = ['*', "x", "o", "."]
COLORS = []
