dapr uninstall --all
dapr init

echo "export PATH=$PATH:/home/vscode/.dapr/bin" >> ~/.zshrc
echo "export PATH=$PATH:/home/vscode/.dapr/bin" >> ~/.zprofile
echo "eval '$(starship init zsh)'" >> ~/.zshrc