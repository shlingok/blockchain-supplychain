package main

import (
    "github.com/hyperledger/fabric/core/chaincode/shim"
    "fmt"
)

const ALL_ORDER_IDS_KEY = "ORDERS"

const STATUS_TYPE_SOURCE = "SOURCE"

const STATUS_TYPE_TRANSPORT = "TRANSPORT"

//==============================================================================================================================
//	 Invocations
//==============================================================================================================================
func AddOrder(stub shim.ChaincodeStubInterface, callerDetails CallerDetails, order Order) (error) {
    //TODO Validate order

    if callerDetails.Role != ROLE_INTERNAL_SYSTEM { return LogAndError("Caller does not have permission to add an order")}

    //Add to state
    _, err := SaveOrder(stub, order)
    if err != nil { return err }

    err = addOrderIdToHolder(stub, order.Id)
    if err != nil { return LogAndError(err.Error()) }

    //Add blank order history
    orderHistory := OrderHistory{[]OrderUpdate{}}
    _, err = SaveOrderHistory(stub, ORDER_HISTORY_KEY_PREFIX + order.Id, orderHistory)

    return err
}

func UpdateOrderStatus(stub shim.ChaincodeStubInterface, callerDetails CallerDetails, orderId string, statusType string, statusValue string, comment string) (error) {
    //TODO Validate update (define state progression)
    //TODO permissions
    order, err := RetrieveOrder(stub, orderId)

    if err != nil { return LogAndError(err.Error()) }

    //TODO make this more extendable
    fromValue := ""
    updateType := ""
    if statusType == STATUS_TYPE_SOURCE {
        updateType = UPDATE_TYPE_SOURCE_STATUS
        fromValue = order.Source.Status
        order.Source.Status = statusValue
    } else if statusType == STATUS_TYPE_TRANSPORT {
        updateType = UPDATE_TYPE_TRANSPORT_STATUS
        fromValue = order.Transport.Status
        order.Transport.Status = statusValue
    }

    //Add to state
    _, err = SaveOrder(stub, order)
    if err != nil { return LogAndError(err.Error()) }

    timestamp, err := stub.GetTxTimestamp()
    if err != nil { return LogAndError(err.Error()) }

    //Update history
    orderUpdate := NewOrderUpdate(updateType, fromValue, statusValue, comment, callerDetails.Username, timestamp.String())

    return UpdateOrderHistory(stub, orderId, orderUpdate)
}
//==============================================================================================================================
//	 Queries
//==============================================================================================================================

func GetOrder(stub shim.ChaincodeStubInterface, callerDetails CallerDetails, orderId string) (Order, bool, error) {

    //TODO Has permission to retrieve?

    order, err := RetrieveOrder(stub, orderId)

    return order, false, err
}

func GetAllOrders(stub shim.ChaincodeStubInterface, callerDetails CallerDetails) (Orders, error) {

    //TODO Has permission to retrieve?
    orders := Orders{}

    orderIds, err := RetrieveIdsHolder(stub, ALL_ORDER_IDS_KEY)

    if err != nil { return orders, LogAndError("Unable to retrieve order id holder") }

    for _, orderId := range orderIds.Ids {
        order, accessDenied, err := GetOrder(stub, callerDetails, orderId)

        if accessDenied {
            fmt.Println("Access denied when reading order: " + orderId)
        } else if (err != nil) {
            return orders, LogAndError("There was an error when retrieving order: " + err.Error())
        } else {
            orders.Orders = append(orders.Orders, order)
        }
    }

    return orders, err
}

//==============================================================================================================================
//	 Internal
//==============================================================================================================================

func addOrderIdToHolder(stub shim.ChaincodeStubInterface, orderId string) (error) {
    idHolder, err := RetrieveIdsHolder(stub, ALL_ORDER_IDS_KEY)

    if err != nil {
        fmt.Println("Unable to retrieve id holder so this is probably the first order...adding")

        idHolder = IdsHolder{}
    }

    idHolder.Ids = append(idHolder.Ids, orderId)

    _, err = SaveIdsHolder(stub, ALL_ORDER_IDS_KEY, idHolder)

    return err
}