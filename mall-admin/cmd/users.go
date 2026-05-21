package cmd

import (
	"fmt"
	"mall-admin/internal/models"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage platform users",
	Long:  "Commands for listing, searching, viewing, and managing platform user accounts.",
}

// users list ----------------------------------------------------------------

var (
	usersListPage int
	usersListAll  bool
)

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users (paginated, 20 per page)",
	RunE: func(cmd *cobra.Command, args []string) error {
		var users []models.SysUser
		q := DB.Unscoped()
		if !usersListAll {
			offset := (usersListPage - 1) * 20
			q = q.Offset(offset).Limit(20)
		}
		if err := q.Find(&users).Error; err != nil {
			return fmt.Errorf("query failed: %w", err)
		}
		printUserTable(users)
		return nil
	},
}

// users search ---------------------------------------------------------------

var (
	searchUsername string
	searchPhone    string
	searchNickname string
	searchPage     int
)

var usersSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search users by username, phone, or nickname",
	RunE: func(cmd *cobra.Command, args []string) error {
		if searchUsername == "" && searchPhone == "" && searchNickname == "" {
			return fmt.Errorf("at least one filter flag is required (--username, --phone, --nickname)")
		}
		q := DB.Unscoped()
		if searchUsername != "" {
			q = q.Where("username LIKE ?", "%"+searchUsername+"%")
		}
		if searchPhone != "" {
			q = q.Where("phone LIKE ?", "%"+searchPhone+"%")
		}
		if searchNickname != "" {
			q = q.Where("nick_name LIKE ?", "%"+searchNickname+"%")
		}
		var users []models.SysUser
		offset := (searchPage - 1) * 20
		if err := q.Offset(offset).Limit(20).Find(&users).Error; err != nil {
			return fmt.Errorf("query failed: %w", err)
		}
		printUserTable(users)
		return nil
	},
}

// users show -----------------------------------------------------------------

var usersShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show full details for a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		var user models.SysUser
		if err := DB.Unscoped().First(&user, id).Error; err != nil {
			fmt.Fprintln(os.Stderr, "User not found.")
			os.Exit(1)
		}
		printUserDetail(user)
		return nil
	},
}

// users set-role -------------------------------------------------------------

var usersSetRoleCmd = &cobra.Command{
	Use:   "set-role <id> <role>",
	Short: "Update the role of a user",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		newRole, parseErr := parseStatus(args[1]) // reuse int parser (non-negative check differs but 0–2 is fine for now)
		if parseErr != nil {
			// Allow any non-negative integer for role
			var n int
			if _, scanErr := fmt.Sscan(args[1], &n); scanErr != nil || n < 0 {
				return fmt.Errorf("invalid role %q: must be a non-negative integer", args[1])
			}
			newRole = n
		}

		var user models.SysUser
		if err := DB.Unscoped().First(&user, id).Error; err != nil {
			fmt.Fprintln(os.Stderr, "User not found.")
			os.Exit(1)
		}

		prompt := fmt.Sprintf("Update role for user #%d (%s) from %d → %d?", user.ID, user.Username, user.RoleID, newRole)
		if !mustConfirm(prompt) {
			fmt.Println("Aborted.")
			return nil
		}

		if err := DB.Model(&user).Update("role_id", newRole).Error; err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		fmt.Printf("✓ User #%d role updated to %d.\n", user.ID, newRole)
		return nil
	},
}

// users delete ---------------------------------------------------------------

var usersDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Soft-delete a user account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}

		var user models.SysUser
		if err := DB.Unscoped().First(&user, id).Error; err != nil {
			fmt.Fprintln(os.Stderr, "User not found.")
			os.Exit(1)
		}

		prompt := fmt.Sprintf("Soft-delete user #%d (username: %s)?", user.ID, user.Username)
		if !mustConfirm(prompt) {
			fmt.Println("Aborted.")
			return nil
		}

		if err := DB.Delete(&user).Error; err != nil {
			return fmt.Errorf("delete failed: %w", err)
		}
		fmt.Printf("✓ User #%d deleted.\n", user.ID)
		return nil
	},
}

// helpers --------------------------------------------------------------------

func printUserTable(users []models.SysUser) {
	if len(users) == 0 {
		fmt.Println("No records found.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tUsername\tNickname\tPhone\tEmail\tRole\tCreated\tStatus")
	fmt.Fprintln(w, "--\t--------\t--------\t-----\t-----\t----\t-------\t------")
	for _, u := range users {
		status := ""
		if u.DeletedAt.Valid {
			status = "[DELETED]"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			u.ID, u.Username, u.NickName, models.MaskPhone(u.Phone),
			u.Email, u.RoleID, u.CreatedAt.Format("2006-01-02"), status)
	}
	w.Flush()
}

func printUserDetail(u models.SysUser) {
	deletedAt := "-"
	if u.DeletedAt.Valid {
		deletedAt = u.DeletedAt.Time.Format(time.RFC3339)
	}
	fmt.Printf("ID:          %d\n", u.ID)
	fmt.Printf("Username:    %s\n", u.Username)
	fmt.Printf("Nickname:    %s\n", u.NickName)
	fmt.Printf("Phone:       %s\n", u.Phone)
	fmt.Printf("Email:       %s\n", u.Email)
	fmt.Printf("Sex:         %s\n", u.Sex)
	fmt.Printf("Avatar:      %s\n", u.Avatar)
	fmt.Printf("WxUserid:    %s\n", u.WxUserid)
	fmt.Printf("WxOpenid:    %s\n", u.WxOpenid)
	fmt.Printf("WxUnionid:   %s\n", u.WxUnionid)
	fmt.Printf("Remarks:     %s\n", u.Remarks)
	fmt.Printf("RoleID:      %d\n", u.RoleID)
	fmt.Printf("CreatedAt:   %s\n", u.CreatedAt.Format(time.RFC3339))
	fmt.Printf("UpdatedAt:   %s\n", u.UpdatedAt.Format(time.RFC3339))
	fmt.Printf("DeletedAt:   %s\n", deletedAt)
}

func init() {
	// list flags
	usersListCmd.Flags().IntVar(&usersListPage, "page", 1, "page number (20 rows per page)")
	usersListCmd.Flags().BoolVar(&usersListAll, "all", false, "show all records without pagination")

	// search flags
	usersSearchCmd.Flags().StringVar(&searchUsername, "username", "", "filter by username (partial match)")
	usersSearchCmd.Flags().StringVar(&searchPhone, "phone", "", "filter by phone number (partial match)")
	usersSearchCmd.Flags().StringVar(&searchNickname, "nickname", "", "filter by nickname (partial match)")
	usersSearchCmd.Flags().IntVar(&searchPage, "page", 1, "page number")

	usersCmd.AddCommand(usersListCmd, usersSearchCmd, usersShowCmd, usersSetRoleCmd, usersDeleteCmd)
	rootCmd.AddCommand(usersCmd)
}
