package purchase

import "fmt"

// NeedsLicense determines whether a license is needed to drive a type of vehicle. Only "car" and "truck" require a license.
func NeedsLicense(kind string) bool {
	// panic("NeedsLicense not implemented")
    switch {
        case kind == "car":
        	return true
        case kind == "truck":
        	return true
        default :
      	 	return false
    }	
}

// ChooseVehicle recommends a vehicle for selection. It always recommends the vehicle that comes first in lexicographical order.
func ChooseVehicle(option1, option2 string) string {
	// panic("ChooseVehicle not implemented")
     switch {
        case option1 > option2 :
        	return fmt.Sprintf("%s is clearly the better choice.", option2)
        case option1 < option2:
        	return fmt.Sprintf("%s is clearly the better choice.", option1)
        default :
      	 	return ""
    }	
}

// CalculateResellPrice calculates how much a vehicle can resell for at a certain age.
func CalculateResellPrice(originalPrice, age float64) float64 {
	// panic("CalculateResellPrice not implemented")
    var price float64
    switch {
        case age < 3 :
            price = originalPrice * 0.8
        case age >= 10:
            price = originalPrice * 0.5
        case age >=3 && age < 10:
            price = originalPrice * 0.7
    }
    return price
}
