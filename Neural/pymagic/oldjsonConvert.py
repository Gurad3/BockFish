import json


with open("TrainedMNIST-OldStruct.json", 'r') as file:
    data = json.load(file)



newW = []
w = data["w"]

for x in range(len(w)):
    newA = []

    for b in range(len(w[x])):

        for w1 in range(len(w[x][b])):
            newA.append(w[x][b][w1])

    newW.append(newA)

newData = {"layer_biases":data["b"], "layer_weights": newW}


with open("test.json", 'w') as file2:
    json.dump(newData, file2)