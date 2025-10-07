package internal

import (
	pb "github.com/pongsathonn/ihavefood/src/merchantservice/genproto"
)

type NewMerchant struct {
	Name        string         `json:"merchantName"`
	Menu        []*NewMenuItem `json:"menu"`
	ImageInfo   *DbImageInfo   `json:"imageInfo"`
	Address     *NewAddress    `json:"address,omitempty"`
	PhoneNumber string         `json:"phoneNumber,omitempty"`
	Status      string         `json:"status,omitempty"`
}

type NewMenuItem struct {
	FoodName  string       `json:"foodName"`
	Price     int32        `json:"price"`
	ImageInfo *DbImageInfo `json:"imageInfo"`
}

type NewAddress struct {
	AddressName string `json:"addressName"`
	SubDistrict string `json:"subDistrict"`
	District    string `json:"district"`
	Province    string `json:"province"`
	PostalCode  string `json:"postalCode"`
}

type DbMerchant struct {
	ID          string        `bson:"_id,omitempty"`
	Name        string        `bson:"name"`
	Menu        []*DbMenuItem `bson:"menu"`
	ImageInfo   *DbImageInfo  `bson:"imageInfo"`
	Address     *DbAddress    `bson:"address,omitempty"`
	PhoneNumber string        `bson:"phoneNumber,omitempty"`
	Status      string        `bson:"status"`
}

type DbMenuItem struct {
	ItemID    string       `bson:"item_id"`
	FoodName  string       `bson:"foodName"`
	Price     int32        `bson:"price"`
	ImageInfo *DbImageInfo `bson:"imageInfo"`
}

type DbImageInfo struct {
	Url  string `bson:"url"`
	Type string `bson:"type"`
}

type DbAddress struct {
	AddressID   string `bson:"address_id"`
	AddressName string `bson:"addressName"`
	SubDistrict string `bson:"subDistrict"`
	District    string `bson:"district"`
	Province    string `bson:"province"`
	PostalCode  string `bson:"postalCode"`
}

func (nm *NewMerchant) FromProto(req *pb.CreateMerchantRequest) *NewMerchant {
	if req == nil {
		return nil
	}

	menuItems := make([]*NewMenuItem, 0, len(req.Menu))
	for _, m := range req.Menu {
		menuItems = append(menuItems, &NewMenuItem{
			FoodName: m.FoodName,
			Price:    m.Price,
			ImageInfo: &DbImageInfo{
				Url:  m.Image.Url,
				Type: m.Image.Type,
			},
		})
	}

	var addr *NewAddress
	if req.Address != nil {
		addr = &NewAddress{
			AddressName: req.Address.AddressName,
			SubDistrict: req.Address.SubDistrict,
			District:    req.Address.District,
			Province:    req.Address.Province,
			PostalCode:  req.Address.PostalCode,
		}
	}

	return &NewMerchant{
		Name: req.MerchantName,
		Menu: menuItems,
		ImageInfo: &DbImageInfo{
			Url:  req.Image.Url,
			Type: req.Image.Type,
		},
		Address:     addr,
		PhoneNumber: req.Phone,
		Status:      req.Status.String(),
	}
}

func (dm *DbMerchant) IntoProto() *pb.Merchant {
	if dm == nil {
		return nil
	}

	menuItems := make([]*pb.MenuItem, 0, len(dm.Menu))
	for _, m := range dm.Menu {
		menuItems = append(menuItems, &pb.MenuItem{
			ItemId:   m.ItemID,
			FoodName: m.FoodName,
			Price:    m.Price,
			Image: &pb.ImageInfo{
				Url:  m.ImageInfo.Url,
				Type: m.ImageInfo.Url,
			},
		})
	}

	var addr *pb.Address
	if dm.Address != nil {
		addr = &pb.Address{
			AddressId:   dm.Address.AddressID,
			AddressName: dm.Address.AddressName,
			SubDistrict: dm.Address.SubDistrict,
			District:    dm.Address.District,
			Province:    dm.Address.Province,
			PostalCode:  dm.Address.PostalCode,
		}
	}

	return &pb.Merchant{
		MerchantId:   dm.ID,
		MerchantName: dm.Name,
		Menu:         menuItems,
		Address:      addr,
		Phone:        dm.PhoneNumber,
		Image: &pb.ImageInfo{
			Url:  dm.ImageInfo.Url,
			Type: dm.ImageInfo.Type,
		},
		Status: pb.StoreStatus(pb.StoreStatus_value[dm.Status]),
	}
}
