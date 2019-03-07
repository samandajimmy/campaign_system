package usecase

import (
	"context"
	"fmt"
	"gade/srv-gade-point/campaigns"
	"gade/srv-gade-point/models"
	"gade/srv-gade-point/vouchers"
	"io"
	"math"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"
)

const (
	letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	lengthCode  = 5
	timeFormat  = "2006-01-02T15:04:05.999Z07:00" // reduce precision from RFC3339Nano as date format
)

var (
	statusVoucher = []string{"0", "1"} // voucher status
	floatType     = reflect.TypeOf(float64(0))
)

type voucherUseCase struct {
	voucherRepo    vouchers.Repository
	campaignRepo   campaigns.Repository
	contextTimeout time.Duration
}

// NewVoucherUseCase will create new an voucherUseCase object representation of vouchers.UseCase interface
func NewVoucherUseCase(vchrRepo vouchers.Repository, campgnRepo campaigns.Repository, timeout time.Duration) vouchers.UseCase {
	return &voucherUseCase{
		voucherRepo:    vchrRepo,
		campaignRepo:   campgnRepo,
		contextTimeout: timeout,
	}
}

func (vchr *voucherUseCase) CreateVoucher(c context.Context, m *models.Voucher) error {
	now := time.Now()
	promoCode := make([]*models.PromoCode, 0)
	ctx, cancel := context.WithTimeout(c, vchr.contextTimeout)

	defer cancel()

	err := vchr.voucherRepo.CreateVoucher(ctx, m)
	code, err := generatePromoCode(m.Stock)
	if err != nil {
		return err
	}

	for i := 0; i < len(code); i++ {
		ap := new(models.PromoCode)

		ap = &models.PromoCode{
			PromoCode: m.PrefixPromoCode + code[i],
			Status:    0,
			Voucher:   m,
			CreatedAt: &now,
		}
		promoCode = append(promoCode, ap)
	}
	err = vchr.voucherRepo.CreatePromoCode(ctx, promoCode)
	if err != nil {
		//Delete voucher when failed generate promo code
		err = vchr.voucherRepo.DeleteVoucher(ctx, m.ID)
		if err != nil {
			err = vchr.voucherRepo.DeleteVoucher(ctx, m.ID)
			return err
		}
		return err
	}

	return nil
}

func (vchr *voucherUseCase) UpdateVoucher(c context.Context, id int64, updateVoucher *models.UpdateVoucher) error {

	ctx, cancel := context.WithTimeout(c, vchr.contextTimeout)
	defer cancel()

	err := vchr.voucherRepo.UpdateVoucher(ctx, id, updateVoucher)
	if err != nil {
		return err
	}

	return nil
}

func (vchr *voucherUseCase) UploadVoucherImages(file *multipart.FileHeader) (string, error) {

	src, err := file.Open()
	if err != nil {
		return "", err
	}

	defer src.Close()

	ext := filepath.Ext(file.Filename)
	nsec := time.Now().UnixNano() // number of nanoseconds unix
	fileName := strconv.FormatInt(nsec, 10) + ext

	filePathUpload := os.Getenv(`VOUCHER_UPLOAD_PATH`) + fileName
	filePathPublic := os.Getenv(`VOUCHER_PATH`) + "/" + fileName

	dst, err := os.Create(filePathUpload)
	if err != nil {
		return "", err
	}

	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	return filePathPublic, nil
}

func (vchr *voucherUseCase) GetVouchersAdmin(c context.Context, name string, status string, startDate string, endDate string, page int32, limit int32) ([]*models.Voucher, string, error) {
	var listVoucher []*models.Voucher
	var err error
	var totalCount int
	ctx, cancel := context.WithTimeout(c, vchr.contextTimeout)
	defer cancel()

	listVoucher, err = vchr.voucherRepo.GetVouchersAdmin(ctx, name, status, startDate, endDate, page, limit)
	if err != nil {
		return nil, "", err
	}

	totalCount, err = vchr.voucherRepo.CountVouchers(ctx, name, status, startDate, endDate, false)
	if err != nil {
		return nil, "", err
	}

	return listVoucher, strconv.Itoa(totalCount), nil
}

func (vchr *voucherUseCase) GetVoucherAdmin(c context.Context, voucherID string) (*models.Voucher, error) {
	var voucherDetail *models.Voucher
	var err error
	ctx, cancel := context.WithTimeout(c, vchr.contextTimeout)
	defer cancel()

	voucherDetail, err = vchr.voucherRepo.GetVoucherAdmin(ctx, voucherID)
	if err != nil {
		return nil, err
	}

	return voucherDetail, nil
}

func (vchr *voucherUseCase) GetVouchers(c context.Context, name string, status string, startDate string, endDate string, page int32, limit int32) ([]*models.Voucher, string, error) {
	var listVoucher []*models.Voucher
	var err error
	var totalCount int
	ctx, cancel := context.WithTimeout(c, vchr.contextTimeout)
	defer cancel()

	listVoucher, err = vchr.voucherRepo.GetVouchers(ctx, name, startDate, endDate, page, limit)
	if err != nil {
		return nil, "", err
	}

	totalCount, err = vchr.voucherRepo.CountVouchers(ctx, name, statusVoucher[1], startDate, endDate, true)
	if err != nil {
		return nil, "", err
	}

	return listVoucher, strconv.Itoa(totalCount), nil
}

func (vchr *voucherUseCase) GetVoucher(c context.Context, voucherID string) (*models.Voucher, error) {
	var voucherDetail *models.Voucher
	var err error
	ctx, cancel := context.WithTimeout(c, vchr.contextTimeout)
	defer cancel()

	voucherDetail, err = vchr.voucherRepo.GetVoucher(ctx, voucherID)
	if err != nil {
		return nil, err
	}

	return voucherDetail, nil
}

func (vchr *voucherUseCase) GetVouchersUser(c context.Context, userID string, status string, page int32, limit int32) ([]models.PromoCode, string, error) {
	var err error
	var totalCount int
	ctx, cancel := context.WithTimeout(c, vchr.contextTimeout)
	defer cancel()

	vouchersUser, err := vchr.voucherRepo.GetVouchersUser(ctx, userID, status, page, limit)
	if err != nil {
		return nil, "", err
	}

	totalCount, err = vchr.voucherRepo.CountPromoCode(ctx, status, userID)
	if err != nil {
		return nil, "", err
	}

	return vouchersUser, strconv.Itoa(totalCount), nil
}

func (vchr *voucherUseCase) VoucherBuy(c context.Context, m *models.PayloadVoucherBuy) (*models.PromoCode, error) {
	var err error
	now := time.Now()
	c, cancel := context.WithTimeout(c, vchr.contextTimeout)
	defer cancel()

	err = vchr.voucherRepo.VoucherCheckExpired(c, m.VoucherID)
	if err != nil {
		return nil, err
	}

	voucherDetail, err := vchr.voucherRepo.GetVoucher(c, m.VoucherID)
	if err != nil {
		return nil, err
	}

	userPoint, err := vchr.campaignRepo.GetUserPoint(c, m.UserID)
	if err != nil {
		return nil, err
	}

	err = validateBuy(voucherDetail.Point, int64(userPoint), voucherDetail.Available)
	if err != nil {
		return nil, err
	}

	promoCode, err := vchr.voucherRepo.UpdatePromoCodeBought(c, m.VoucherID, m.UserID)
	if err != nil {
		return nil, err
	}

	// Parse interface to float
	parseFloat, err := getFloat(voucherDetail.Point)
	pointAmount := math.Floor(parseFloat)

	campaignTrx := &models.CampaignTrx{
		UserID:          m.UserID,
		PointAmount:     &pointAmount,
		TransactionType: models.TransactionPointTypeKredit,
		TransactionDate: &now,
		PromoCode:       promoCode,
		CreatedAt:       &now,
	}

	err = vchr.campaignRepo.SavePoint(c, campaignTrx)
	if err != nil {
		return nil, err
	}

	promoCode.Voucher = voucherDetail

	return promoCode, nil
}

func (vchr *voucherUseCase) VoucherValidate(c context.Context, validateVoucher *models.PayloadValidateVoucher) (*models.Voucher, error) {
	var err error

	c, cancel := context.WithTimeout(c, vchr.contextTimeout)
	defer cancel()

	err = vchr.voucherRepo.VoucherCheckExpired(c, validateVoucher.VoucherID)
	if err != nil {
		return nil, err
	}

	voucher, err := vchr.voucherRepo.VoucherCheckMinimalTransaction(c, validateVoucher)
	if err != nil {
		return nil, err
	}

	return voucher, nil
}

func (vchr *voucherUseCase) VoucherRedeem(c context.Context, voucherRedeem *models.PayloadValidateVoucher) (*models.PromoCode, error) {
	var err error

	c, cancel := context.WithTimeout(c, vchr.contextTimeout)
	defer cancel()

	promoCode, err := vchr.voucherRepo.UpdatePromoCodeRedeemed(c, voucherRedeem.VoucherID, voucherRedeem.UserID)
	if err != nil {
		return nil, err
	}

	return promoCode, nil
}

func generatePromoCode(stock int32) (code []string, err error) {

	var arr = make([]string, stock)
	for i := range arr {
		arr[i] = randStringBytes(lengthCode)
	}

	return arr, nil
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func validateBuy(voucherPoint int64, userPoint int64, avaliable *int32) error {
	if *avaliable <= 0 {
		return models.ErrVoucherUnavailable
	}
	if userPoint < voucherPoint {
		return models.ErrPointDeficit
	}
	return nil
}

func getFloat(unk interface{}) (float64, error) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}
