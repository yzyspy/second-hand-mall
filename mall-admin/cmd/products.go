package cmd

import (
	"fmt"
	"mall-admin/internal/models"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var productsCmd = &cobra.Command{
	Use:   "products",
	Short: "Manage product listings",
	Long:  "Commands for listing, searching, viewing, and managing product listings.",
}

// products list --------------------------------------------------------------

var (
	productsListStatus int
	productsListUserID uint
	productsListPage   int
	productsListAll    bool
)

var productsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all products (paginated, 20 per page)",
	RunE: func(cmd *cobra.Command, args []string) error {
		q := DB.Unscoped()
		if cmd.Flags().Changed("status") {
			q = q.Where("status = ?", productsListStatus)
		}
		if cmd.Flags().Changed("user-id") {
			q = q.Where("user_id = ?", productsListUserID)
		}
		var products []models.Product
		if !productsListAll {
			offset := (productsListPage - 1) * 20
			q = q.Offset(offset).Limit(20)
		}
		if err := q.Find(&products).Error; err != nil {
			return fmt.Errorf("query failed: %w", err)
		}
		printProductTable(products)
		return nil
	},
}

// products show --------------------------------------------------------------

var productsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show full details for a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		var product models.Product
		if err := DB.Unscoped().First(&product, id).Error; err != nil {
			fmt.Fprintln(os.Stderr, "Product not found.")
			os.Exit(1)
		}
		var favCount int64
		DB.Model(&models.UserFavorite{}).Where("product_id = ?", product.ID).Count(&favCount)
		printProductDetail(product, favCount)
		return nil
	},
}

// products set-status --------------------------------------------------------

var productsSetStatusCmd = &cobra.Command{
	Use:   "set-status <id> <status>",
	Short: "Update the status of a product (0=available, 1=sold, 2=removed)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		newStatus, err := parseStatus(args[1])
		if err != nil {
			return err
		}

		var product models.Product
		if err := DB.Unscoped().First(&product, id).Error; err != nil {
			fmt.Fprintln(os.Stderr, "Product not found.")
			os.Exit(1)
		}

		prompt := fmt.Sprintf("Change product #%d status from %s → %s?",
			product.ID, models.StatusLabel(product.Status), models.StatusLabel(newStatus))
		if !mustConfirm(prompt) {
			fmt.Println("Aborted.")
			return nil
		}

		if err := DB.Model(&product).Update("status", newStatus).Error; err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		fmt.Printf("✓ Product #%d status updated to %s.\n", product.ID, models.StatusLabel(newStatus))
		return nil
	},
}

// products delete ------------------------------------------------------------

var productsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Soft-delete a product listing",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}

		var product models.Product
		if err := DB.Unscoped().First(&product, id).Error; err != nil {
			fmt.Fprintln(os.Stderr, "Product not found.")
			os.Exit(1)
		}

		prompt := fmt.Sprintf("Soft-delete product #%d (%q)?", product.ID, product.Title)
		if !mustConfirm(prompt) {
			fmt.Println("Aborted.")
			return nil
		}

		if err := DB.Delete(&product).Error; err != nil {
			return fmt.Errorf("delete failed: %w", err)
		}
		fmt.Printf("✓ Product #%d deleted.\n", product.ID)
		return nil
	},
}

// helpers --------------------------------------------------------------------

func printProductTable(products []models.Product) {
	if len(products) == 0 {
		fmt.Println("No records found.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTitle\tPrice\tStatus\tSeller\tCreated\tDeleted")
	fmt.Fprintln(w, "--\t-----\t-----\t------\t------\t-------\t-------")
	for _, p := range products {
		deleted := ""
		if p.DeletedAt.Valid {
			deleted = "[DELETED]"
		}
		title := p.Title
		if len(title) > 30 {
			title = title[:27] + "..."
		}
		fmt.Fprintf(w, "%d\t%s\t%.2f\t%s\t%d\t%s\t%s\n",
			p.ID, title, p.Price, models.StatusLabel(p.Status),
			p.UserID, p.CreatedAt.Format("2006-01-02"), deleted)
	}
	w.Flush()
}

func printProductDetail(p models.Product, favorites int64) {
	deletedAt := "-"
	if p.DeletedAt.Valid {
		deletedAt = p.DeletedAt.Time.Format(time.RFC3339)
	}
	fmt.Printf("ID:           %d\n", p.ID)
	fmt.Printf("Title:        %s\n", p.Title)
	fmt.Printf("Price:        %.2f\n", p.Price)
	fmt.Printf("Status:       %s (%d)\n", models.StatusLabel(p.Status), p.Status)
	fmt.Printf("Location:     %s\n", p.Location)
	fmt.Printf("Seller (uid): %d\n", p.UserID)
	fmt.Printf("Buyer (uid):  %d\n", p.BuyUID)
	fmt.Printf("ContactType:  %s\n", p.ContactType)
	fmt.Printf("ContactValue: %s\n", p.ContactValue)
	fmt.Printf("Images:       %s\n", p.Images)
	fmt.Printf("Description:\n%s\n", p.Description)
	fmt.Printf("Favorites:    %d\n", favorites)
	fmt.Printf("CreatedAt:    %s\n", p.CreatedAt.Format(time.RFC3339))
	fmt.Printf("UpdatedAt:    %s\n", p.UpdatedAt.Format(time.RFC3339))
	fmt.Printf("DeletedAt:    %s\n", deletedAt)
}

func init() {
	productsListCmd.Flags().IntVar(&productsListStatus, "status", 0, "filter by status (0=available, 1=sold, 2=removed)")
	productsListCmd.Flags().UintVar(&productsListUserID, "user-id", 0, "filter by seller user id")
	productsListCmd.Flags().IntVar(&productsListPage, "page", 1, "page number (20 rows per page)")
	productsListCmd.Flags().BoolVar(&productsListAll, "all", false, "show all records without pagination")

	productsCmd.AddCommand(productsListCmd, productsShowCmd, productsSetStatusCmd, productsDeleteCmd)
	rootCmd.AddCommand(productsCmd)
}
