#!/usr/bin/env fish

# Configuration: Update these paths!
set DOTFILES_DIR "$HOME/dotfiles"
set CONFIG_DIR "$HOME/.config"

# Folders/Files you want to track (add yours here)
set TARGETS hypr foot kitty fastfetch alacritty 

echo "--- 🚀 Syncing Dotfiles ---"

# Ensure the dotfiles directory exists
mkdir -p $DOTFILES_DIR

# Copy latest configs to your git repo folder
for target in $TARGETS
    if test -d "$CONFIG_DIR/$target" -o -f "$CONFIG_DIR/$target"
        # Use -L to follow symlinks and copy actual content
        cp -rL "$CONFIG_DIR/$target" "$DOTFILES_DIR/"
        echo "✅ Synced: $target"
    else
        echo "⚠️  Skipped: $target (not found)"
    end
end

# Git Automation
cd $DOTFILES_DIR
if test -d .git
    git add .
    git commit -m "Update rice: "(date +'%Y-%m-%d %H:%M')
    
    # Get current branch name
    set BRANCH (git branch --show-current)
    
    # Push with upstream setup (handles first push automatically)
    if git push -u origin $BRANCH
        echo "--- ☁️  Pushed to Remote ---"
    else
        echo "❌ Error: Failed to push to remote. Check your git remote configuration."
    end
else
    echo "❌ Error: $DOTFILES_DIR is not a git repository."
end
