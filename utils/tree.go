package utils

func GetCameraList(authUser *AuthorizedUser, token *XMLResponse) []NodeBody {
	rootNode := GetTreeViewItem(authUser, token, nil)

	sendParams := &NodeBody{
		ParentId: rootNode[0].Id,
		Id:       GetNextNodeId(rootNode[0].Id),
		NodeType: "Customer",
	}

	firstNode := GetTreeViewItem(authUser, token, sendParams)

	paramList := make([]NodeBody, len(firstNode))
	responseList := make([]NodeBody, 0)
	for i := range firstNode {
		paramList[i] = NodeBody{
			ParentId: firstNode[i].ParentId,
			Id:       GetNextNodeId(firstNode[i].ParentId),
			NodeType: "Site",
		}

		responseList = append(responseList, GetTreeViewItem(authUser, token, &paramList[i])...)
	}

	cameraList := make([]NodeBody, 0)

	for _, device := range responseList {
		if device.EntityType == "CAMERA" {
			cameraList = append(cameraList, device)
		}
	}

	return cameraList
}
