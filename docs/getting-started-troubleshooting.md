# Getting Started Troubleshooting Guide

This guide helps you resolve common issues when getting started with Jot. If you encounter a problem, find the error message or symptom below to get a solution.

---

## Quick Diagnosis

Before diving into specific sections, use this flowchart to identify your issue:

1. **Can't run `jot` command?** → [CLI Issues](#cli-issues)
2. **Getting errors about notebooks?** → [Import Issues](#import-issues)
3. **SQL queries not working?** → [SQL Query Issues](#sql-query-issues)
4. **Searches are slow?** → [Performance Issues](#performance-issues)
5. **Integration with tools failing?** → [Integration Issues](#integration-issues)

---

## Import Issues

### "Notebook not found" Error

**Error message**: `Error: notebook not found` or `config: notebook "MyNotebook" not found`

**Causes and solutions**:

1. **Notebook doesn't exist yet**

   ```bash
   # List existing notebooks
   jot notebook list

   # Create the notebook
   jot notebook create "MyNotebook" --path ~/my-notes
   ```

2. **Notebook name is case-sensitive**

   ```bash
   # ❌ Won't work if notebook is named "MyNotebook"
   jot notebook info mynotebook

   # ✅ Use exact name
   jot notebook info MyNotebook
   ```

3. **Notebook configuration file corrupted**

   ```bash
   # Check config file location
   cat ~/.config/jot/config.json

   # Backup and recreate if needed
   cp ~/.config/jot/config.json ~/.config/jot/config.json.bak
   jot notebook create "MyNotebook" --path ~/my-notes --force
   ```

4. **Using wrong notebook in multi-notebook setup**

   ```bash
   # See all notebooks
   jot notebook list

   # Specify notebook explicitly
   jot -n "MyNotebook" notes list

   # Or set default
   jot notebook set-default "MyNotebook"
   ```

---

### Permission Denied Errors

**Error message**: `permission denied`, `access denied`, or `cannot read file`

**Causes and solutions**:

1. **Notebook directory not readable**

   ```bash
   # Check permissions
   ls -la ~/my-notes

   # Fix if needed (allow owner to read)
   chmod 755 ~/my-notes
   chmod -R 755 ~/my-notes/*
   ```

2. **Running as different user than notebook owner**

   ```bash
   # Check who owns the notebook
   ls -la ~/my-notes

   # Run commands as the same user
   sudo -u username jot notebook create ...

   # Or change ownership
   sudo chown -R $USER:$USER ~/my-notes
   ```

3. **SELinux or AppArmor restrictions (Linux)**

   ```bash
   # Check if SELinux is enabled
   getenforce

   # Temporarily disable for testing
   sudo setenforce 0

   # Or configure policy for jot (consult your admin)
   ```

4. **Windows: Path with special characters**

   ```bash
   # ❌ May fail with spaces or special chars
   jot notebook create "My Notes" --path "C:\Users\Me\My Documents\Notes"

   # ✅ Use quoted paths
   jot notebook create "MyNotes" --path "C:\Users\Me\My Documents\Notes"

   # ✅ Or use forward slashes
   jot notebook create "MyNotes" --path "C:/Users/Me/My Documents/Notes"
   ```

---

### Large Collection Import Timeouts

**Error message**: `timeout`, `taking too long`, or `process killed`

**Causes and solutions**:

1. **Very large collection (1000+ files)**

   ```bash
   # Import works, but may take time on first run
   # Be patient - large collections need time to index
   jot notebook create "BigCollection" --path ~/large-notes

   # Subsequent operations will be faster (cached)
   jot notes list  # First time: may take 30-60 seconds
   jot notes list  # Second time: <1 second
   ```

2. **Network-mounted notebook (slow storage)**

   ```bash
   # Network drives are slower
   # Solution: Copy to local storage first
   cp -r /mnt/network-storage/notes ~/local-notes
   jot notebook create "LocalNotes" --path ~/local-notes
   ```

3. **System running out of memory**

   ```bash
   # Check available memory
   free -h  # Linux
   vm_stat  # macOS

   # If low on memory, close other applications
   # Or process in smaller batches:
   jot notebook create "Batch1" --path ~/notes/a-m
   jot notebook create "Batch2" --path ~/notes/n-z
   ```

4. **Antivirus scanning files during import**

   ```bash
   # Add jot to antivirus exclusions
   # Location depends on antivirus (Windows Defender, etc.)

   # Or temporarily disable for import
   # Consult your antivirus documentation
   ```

---

### Special Character Handling in Filenames

**Error message**: `encoding error`, `invalid filename`, or `mojibake` (garbled characters)

**Causes and solutions**:

1. **UTF-8 characters in filenames**

   ```bash
   # UTF-8 filenames are supported
   # If you see garbled text, check terminal encoding

   # Set UTF-8 locale
   export LC_ALL=en_US.UTF-8
   export LANG=en_US.UTF-8

   # Then run jot
   jot notebook create "MyNotebook" --path ~/my-notes
   ```

2. **Special shell characters in filenames**

   ```bash
   # Files with spaces, quotes, etc. are handled automatically
   # Jot safely escapes these

   # These are fine:
   # - "My Note (v2).md"
   # - "TODO: Buy groceries.md"
   # - "Author's notes.md"

   jot notes list  # Shows all correctly
   ```

3. **Windows: Forbidden characters**

   ```bash
   # Windows forbids these in filenames: < > : " / \ | ? *
   # If files have these, rename before importing

   # Problematic example: "Q1 < Q2 Analysis.md"
   # Rename to: "Q1-LessThan-Q2-Analysis.md"
   ```

---

### Symlink and Nested Structure Issues

**Error message**: `circular symlink`, `too many levels`, or `symlink not followed`

**Causes and solutions**:

1. **Circular symlinks (A→B→A)**

   ```bash
   # Check for circular symlinks
   find ~/my-notes -type l | while read link; do
     echo "$link -> $(readlink $link)"
   done

   # Remove problematic symlinks
   rm ~/my-notes/circular-link
   ```

2. **Symlinks to parent directories**

   ```bash
   # ❌ Problematic
   ln -s ~ ~/my-notes/home-link

   # ✅ Better: Use actual path or copy
   cp ~/important-file ~/my-notes/important-file
   ```

3. **Deeply nested symlinks**

   ```bash
   # Jot resolves symlinks safely
   # But deeply nested structures may be slow

   # Check depth
   find ~/my-notes -type l -exec sh -c 'echo "$1" | tr -cd "/" | wc -c' _ {} \; | sort -n | tail -5

   # Flatten if needed by copying
   find ~/my-notes -type l -exec cp {} {}.copy \;
   ```

---

## SQL Query Issues

### "Query returned no results"

**Symptom**: Query runs without error, but returns empty result

**Causes and solutions**:

1. **File pattern doesn't match any files**

   ```bash
   # ❌ Wrong pattern
   jot notes search --sql "
   SELECT * FROM read_markdown('/wrong/*.md')
   "

   # ✅ Check what files exist first
   jot notes list

   # ✅ Use correct pattern
   jot notes search --sql "
   SELECT * FROM read_markdown('*.md')
   "
   ```

2. **Case-sensitive pattern matching**

   ```bash
   # ❌ Won't find files if case doesn't match
   SELECT * FROM read_markdown()
   WHERE file_path LIKE '%todo.md'  -- Looking for lowercase

   # ✅ Use ILIKE for case-insensitive
   SELECT * FROM read_markdown()
   WHERE file_path ILIKE '%todo.md'
   ```

3. **Search term not in content**

   ```bash
   # ❌ Looking for exact phrase
   SELECT * FROM read_markdown()
   WHERE content LIKE 'exact phrase'

   # ✅ Content has this phrase on its own line
   SELECT * FROM read_markdown()
   WHERE content LIKE '%exact phrase%'
   ```

4. **WHERE condition too restrictive**

   ```bash
   # ❌ Filtering out actual results
   SELECT * FROM read_markdown()
   WHERE md_stats(content).words > 10000  -- Very high threshold

   # ✅ Check distribution first
   SELECT COUNT(*) as notes,
          MAX(md_stats(content).words) as max_words
   FROM read_markdown()
   ```

---

### File Pattern Problems

**Error message**: `no such file`, `pattern invalid`, or `0 results`

**Causes and solutions**:

1. **Absolute vs. relative paths**

   ```bash
   # ❌ These don't work as expected in queries
   SELECT * FROM read_markdown('/home/user/notes/*.md')

   # ✅ Use relative to notebook root
   SELECT * FROM read_markdown('*.md')
   SELECT * FROM read_markdown('subfolder/*.md')
   ```

2. **Forgotten wildcards**

   ```bash
   # ❌ Looking for literal folder name
   SELECT * FROM read_markdown('archive')

   # ✅ Need wildcard for files
   SELECT * FROM read_markdown('archive/*.md')
   ```

3. **Special characters in paths**

   ```bash
   # ❌ May fail with spaces
   SELECT * FROM read_markdown('My Projects/*.md')

   # ✅ Use wildcards with care
   SELECT * FROM read_markdown('%my%project%.md')
   ```

---

### Query Performance Degradation

**Symptom**: First query is fast, but subsequent queries are slow

**Causes and solutions**:

1. **Scanning entire content without LIMIT**

   ```bash
   # ❌ Slow: Reads all content
   SELECT * FROM read_markdown()
   WHERE content LIKE '%keyword%'

   # ✅ Fast: Add LIMIT
   SELECT file_path FROM read_markdown()
   WHERE content LIKE '%keyword%'
   LIMIT 100
   ```

2. **Inefficient JOIN patterns**

   ```bash
   # ❌ May be slow on large sets
   SELECT * FROM read_markdown() a
   JOIN read_markdown() b
   ON a.file_path = b.file_path

   # ✅ Better: Use CTE
   WITH notes AS (
    SELECT * FROM read_markdown()
   )
   SELECT * FROM notes LIMIT 100
   ```

3. **Missing indexes on frequently queried fields**

   ```bash
   # Queries on file_path are fastest
   # Queries on content are slower
   # Optimize by checking file_path first

   # ❌ Content scan on all files
   SELECT * FROM read_markdown()
   WHERE content LIKE '%pattern%'

   # ✅ Filter files first
   SELECT * FROM read_markdown()
   WHERE file_path LIKE '%docs%'
   AND content LIKE '%pattern%'
   ```

---

### Memory Issues with Large Queries

**Error message**: `out of memory`, `segmentation fault`, or `process killed`

**Causes and solutions**:

1. **Queries on very large notebooks (10,000+ files)**

   ```bash
   # ❌ May consume all memory
   SELECT content FROM read_markdown()
   WHERE content IS NOT NULL

   # ✅ Use pagination
   SELECT file_path FROM read_markdown()
   LIMIT 1000 OFFSET 0
   ```

2. **Storing results in memory**

   ```bash
   # ❌ Can't display million-row result
   SELECT * FROM read_markdown()

   # ✅ Export incrementally
   jot notes search --sql "
   SELECT file_path FROM read_markdown()
   LIMIT 1000 OFFSET 0
   " | head -100 > results.txt
   ```

3. **Complex aggregations**

   ```bash
   # ❌ Complex aggregation on large set
   SELECT GROUP_CONCAT(content) FROM read_markdown()

   # ✅ Aggregate metadata instead
   SELECT COUNT(*) as total,
          SUM(md_stats(content).words) as total_words
   FROM read_markdown()
   ```

---

### Timeout Errors

**Error message**: `timeout`, `deadline exceeded`, or `query took too long`

**Causes and solutions**:

1. **Very complex query on large collection**

   ```bash
   # ❌ Complex operations
   SELECT DISTINCT content FROM read_markdown()
   WHERE content LIKE '%a%'
   AND content LIKE '%b%'
   AND content LIKE '%c%'

   # ✅ Simpler operations or pagination
   SELECT DISTINCT file_path FROM read_markdown()
   WHERE file_path LIKE '%docs%'
   LIMIT 1000
   ```

2. **Query blocked by other operations**

   ```bash
   # ❌ Can't run while importing
   # Wait for import to complete
   jot notebook create "Notebook" --path ~/notes  # Taking long

   # In another terminal, queries will wait
   jot notes search --sql "SELECT * FROM read_markdown()"  # Waits
   ```

3. **System under heavy load**

   ```bash
   # Check system load
   top  # Linux
   Activity Monitor  # macOS
   Task Manager  # Windows

   # Close other applications to free resources
   ```

---

## CLI Issues

### Command Not Recognized

**Error message**: `jot: command not found` or `unknown command`

**Causes and solutions**:

1. **jot not in PATH**

   ```bash
   # Check if jot is installed
   which jot

   # If not found, install it
   # See README for installation instructions

   # If installed but not found, add to PATH
   export PATH="$PATH:$HOME/.local/bin"
   ```

2. **Typo in command name**

   ```bash
   # ❌ Wrong command
   jot note list

   # ✅ Correct commands
   jot notes list
   jot notebook list
   ```

3. **Subcommand not available**
   ```bash
   # See all available commands
   jot --help
   jot notes --help
   jot notebook --help
   ```

---

### Configuration File Problems

**Error message**: `config: invalid config` or `config: file not found`

**Causes and solutions**:

1. **Config file corrupted**

   ```bash
   # Check config file
   cat ~/.config/jot/config.json

   # Validate JSON (if you have jq)
   jq empty ~/.config/jot/config.json

   # Backup and regenerate if needed
   cp ~/.config/jot/config.json ~/.config/jot/config.json.bak
   rm ~/.config/jot/config.json
   jot notebook list  # Creates fresh config
   ```

2. **Wrong permissions on config file**

   ```bash
   # Config file should be readable by user
   ls -la ~/.config/jot/config.json

   # Fix permissions if needed
   chmod 644 ~/.config/jot/config.json
   chmod 755 ~/.config/jot
   ```

3. **Config file in wrong location (custom HOME)**

   ```bash
   # Jot looks in $HOME/.config/jot
   echo $HOME
   ls -la ~/.config/jot/

   # If custom config location needed, check docs
   ```

---

### Multi-Notebook Conflicts

**Error message**: `multiple notebooks found` or ambiguous notebook reference

**Causes and solutions**:

1. **Two notebooks with similar names**

   ```bash
   # List all notebooks
   jot notebook list

   # Use full name or set default
   jot -n "Work" notes list
   jot notebook set-default "Work"
   ```

2. **Environment variable overriding settings**

   ```bash
   # Check for JOT environment variables
   env | grep JOT

   # Unset if causing conflicts
   unset JOT_NOTEBOOK

   # Use explicit flag instead
   jot -n "MyNotebook" notes list
   ```

3. **Auto-discovery finding multiple matches**

   ```bash
   # Jot searches parent directories for .jot.json
   # May find multiple if nested

   # Specify explicitly
   jot -n "MyNotebook" notes list

   # Or move to directory containing .jot.json
   cd ~/my-notes
   jot notes list
   ```

---

### Environment Variable Conflicts

**Error message**: Environment variables seem ignored or conflicting

**Causes and solutions**:

1. **PATH not including Jot**

   ```bash
   # Check PATH
   echo $PATH

   # Add Jot location
   export PATH="/home/user/.local/bin:$PATH"

   # Make permanent in ~/.bashrc or ~/.zshrc
   echo 'export PATH="/home/user/.local/bin:$PATH"' >> ~/.bashrc
   ```

2. **HOME variable not set (unusual)**

   ```bash
   # Check HOME
   echo $HOME

   # If empty, set it
   export HOME=/home/username

   # Then jot can find config
   ```

3. **Shell not loading environment**

   ```bash
   # Interactive shell should load env
   bash --login

   # Check ~/.bashrc or ~/.profile are executable
   ls -la ~/.bashrc ~/.profile
   ```

---

### Platform-Specific Issues

#### macOS

**Issue**: "jot cannot be opened because it is not from an identified developer"

Solution:

```bash
# Allow in Security & Privacy
# System Preferences → Security & Privacy → General
# Click "Open Anyway" for jot

# Or bypass from terminal
spctl --add --label "Jot" $(which jot)
```

#### Windows

**Issue**: Scripts show `permission denied` even when executable

Solution:

```bash
# Windows uses different permission model
# Make sure script file is in a writable directory

# Use bash directly
bash ~/scripts/my-script.sh

# Or update execution policy (if using PowerShell)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

#### Linux

**Issue**: "cannot execute: required file not found"

Solution:

```bash
# Check if binary is for your architecture
file $(which jot)

# Should show "x86-64 executable" or similar
# If not, you may have wrong binary for your system

# Download correct version for your architecture
# See README for distribution instructions
```

---

## Performance Issues

### Slow Searches

**Symptom**: `jot notes search` takes 10+ seconds

**Causes and solutions**:

1. **First search on large notebook always slower**

   ```bash
   # First run: slower (loading index)
   # jot notes search "keyword"  # ~5-30 seconds

   # Subsequent runs: much faster (cached)
   # jot notes search "keyword"  # <1 second

   # This is normal behavior
   ```

2. **Very large collection with slow storage**

   ```bash
   # Check notebook size
   du -sh ~/my-notes

   # Slow storage (USB drive, network mount) impacts performance
   # Solution: Copy to SSD
   cp -r /mnt/slow-drive/notes ~/fast-ssd/notes
   jot notebook create "Notes" --path ~/fast-ssd/notes
   ```

3. **Regex patterns can be slow**

   ```bash
   # ❌ Complex regex may be slow
   jot notes search "^[A-Z].*\d{4}$"

   # ✅ Simpler patterns are faster
   jot notes search "word1 word2"
   ```

---

### Large Notebook Queries

**Symptom**: Queries on 10,000+ note notebooks are very slow

**Solutions**:

1. **Use metadata instead of content when possible**

   ```bash
   # ❌ Slow: Reads all content
   jot notes search --sql "
   SELECT * FROM read_markdown()
   "

   # ✅ Fast: Uses metadata only
   jot notes search --sql "
   SELECT file_path FROM read_markdown()
   LIMIT 100
   "
   ```

2. **Filter by path pattern first**

   ```bash
   # ❌ Content scan across entire collection
   SELECT * FROM read_markdown()
   WHERE content LIKE '%important%'

   # ✅ Filter by folder first
   SELECT * FROM read_markdown()
   WHERE file_path LIKE '%archive%'
   AND content LIKE '%important%'
   ```

3. **Split into multiple notebooks**

   ```bash
   # Instead of one 20,000-file notebook
   jot notebook create "Archive" --path ~/notes/archive
   jot notebook create "Active" --path ~/notes/active

   # Queries on each will be faster
   jot -n "Active" notes search --sql "SELECT * FROM read_markdown()"
   ```

---

### Complex Query Performance

**Symptom**: Complex SQL queries timeout or run very slowly

**Solutions**:

1. **Break into simpler queries**

   ```bash
   # ❌ Very complex single query
   SELECT complex, nested, aggregate(functions) FROM ...
   WHERE multiple AND conditions AND apply
   GROUP BY multiple fields
   HAVING complex aggregate conditions

   # ✅ Two simpler queries
   jot notes search --sql "SELECT * FROM read_markdown() LIMIT 100"
   # Process results in script or jq
   ```

2. **Use LIMIT during development**

   ```bash
   # Test with small result set first
   SELECT * FROM read_markdown() LIMIT 10

   # Then remove LIMIT when confirmed
   SELECT * FROM read_markdown()
   ```

---

### Database Optimization

**Solutions** for persistent slow performance:

1. **Rebuild notebook index**

   ```bash
   # Remove and recreate notebook
   jot notebook delete "MyNotebook"
   jot notebook create "MyNotebook" --path ~/my-notes
   ```

2. **Check disk space**

   ```bash
   # Insufficient disk space can slow queries
   df -h ~

   # Free up space if needed
   ```

3. **Check for corrupted files**

   ```bash
   # Find files with encoding issues
   find ~/my-notes -name "*.md" -exec file {} \; | grep -v UTF-8

   # Convert if needed
   iconv -f iso-8859-1 -t utf-8 file.md > file-utf8.md
   ```

---

## Integration Issues

### jq Pipeline Failures

**Error message**: `jq: command not found` or JSON parsing error

**Causes and solutions**:

1. **jq not installed**

   ```bash
   # Install jq
   # Ubuntu/Debian
   sudo apt-get install jq

   # macOS
   brew install jq

   # Or download from jqlang.github.io/jq
   ```

2. **Invalid JSON from Jot**

   ```bash
   # Check if jot output is valid JSON
   jot notes search --sql "SELECT * FROM read_markdown() LIMIT 1" | jq .

   # If error, may need --json flag (if available)
   ```

3. **jq syntax error**

   ```bash
   # ❌ Invalid jq syntax
   jot notes search --sql "..." | jq '.[] | invalid syntax'

   # ✅ Valid jq syntax
   jot notes search --sql "..." | jq '.[] | .field'
   ```

---

### Shell Script Compatibility

**Error message**: Script fails with `bad interpreter` or `command not found`

**Causes and solutions**:

1. **Wrong shebang line**

   ```bash
   # ❌ Wrong
   #!/bin/sh  # May not have all bash features

   # ✅ Correct
   #!/bin/bash
   ```

2. **Bash-specific features in /bin/sh script**

   ```bash
   # ❌ Won't work in /bin/sh
   #!/bin/sh
   jot notes search --sql "SELECT ..."  # Command substitution

   # ✅ Fix by using #!/bin/bash
   #!/bin/bash
   ```

3. **Script not executable**

   ```bash
   # Make executable
   chmod +x script.sh

   # Then run
   ./script.sh
   ```

---

### Path Issues on Windows

**Error message**: `file not found`, `path does not exist`, or `invalid path`

**Causes and solutions**:

1. **Backslashes vs. forward slashes**

   ```bash
   # ❌ Backslashes in bash/PowerShell
   jot notebook create "Notes" --path "C:\Users\Me\Notes"

   # ✅ Use forward slashes
   jot notebook create "Notes" --path "C:/Users/Me/Notes"

   # ✅ Or escape backslashes
   jot notebook create "Notes" --path "C:\\Users\\Me\\Notes"
   ```

2. **Spaces in paths**

   ```bash
   # ❌ Unquoted path with spaces fails
   jot notebook create "Notes" --path C:\Users\My Documents\Notes

   # ✅ Quote the path
   jot notebook create "Notes" --path "C:/Users/My Documents/Notes"
   ```

3. **UNC paths (network shares)**

   ```bash
   # UNC paths are supported
   jot notebook create "Remote" --path "//server/share/notes"

   # But may be slower than local storage
   ```

---

### Encoding Problems

**Symptom**: Garbled characters, mojibake, or corruption in search results

**Causes and solutions**:

1. **Terminal encoding mismatch**

   ```bash
   # Check terminal encoding
   echo $LANG

   # Set to UTF-8
   export LC_ALL=en_US.UTF-8
   export LANG=en_US.UTF-8
   ```

2. **File encoding not UTF-8**

   ```bash
   # Check encoding
   file ~/my-notes/*.md

   # Convert to UTF-8
   iconv -f iso-8859-1 -t UTF-8 oldfile.md > newfile.md
   ```

3. **Git/editor saved with wrong encoding**

   ```bash
   # Resave file in UTF-8
   # In most editors: File → Save with Encoding → UTF-8

   # Or use command line
   iconv -f cp1252 -t utf-8 file.md > file-utf8.md
   mv file-utf8.md file.md
   ```

---

### Tool Version Conflicts

**Error message**: `requires version X.Y.Z or higher` or incompatible version

**Causes and solutions**:

1. **Jot version too old**

   ```bash
   # Check version
   jot --version

   # Update to latest
   # See README for update instructions
   ```

2. **DuckDB version mismatch**

   ```bash
   # Jot includes correct DuckDB version
   # Shouldn't need separate installation

   # If issues, try rebuilding
   # See README build instructions
   ```

3. **jq version incompatibility**

   ```bash
   # Check jq version
   jq --version

   # Update jq if needed
   brew upgrade jq  # macOS
   sudo apt-get upgrade jq  # Linux
   ```

---

## FAQ: Common Questions

### Q: How do I get better performance?

**A**:

1. **Use metadata queries** instead of full content scans
2. **Add LIMIT** to queries during development
3. **Filter by path** before filtering by content
4. **Store notebooks on fast storage** (SSD, not network drives)
5. **Split large collections** into multiple notebooks

### Q: Can I use Jot with Obsidian/Bear/Notion?

**A**:

- **Obsidian**: Yes! Jot can read Obsidian vaults directly
- **Bear**: You must export from Bear first (as Markdown)
- **Notion**: Export to Markdown, then import
- See [Import Workflow Guide](import-workflow-guide.md) for details

### Q: Is my data safe with Jot?

**A**:

- **Jot doesn't modify notes** unless you explicitly run commands
- **SQL queries are read-only** (defensive programming prevents writes)
- **Backups recommended** - store notebook on git for version control
- **Local-only by default** - notes stay on your computer

### Q: Can I share notebooks with my team?

**A**:

- **Local notebooks**: Share the folder (via cloud sync, git, etc.)
- **Set up on each machine**: `jot notebook create "Shared" --path /path/to/notes`
- **Git workflow**: Pull latest, run queries, commit results
- **See Notebook Discovery guide** for multi-notebook setups

### Q: How do I export results?

**A**:

- **SQL queries return JSON** - pipe to jq for transformation
- **Export to CSV**: Use SQL's text output or jq conversion
- **Save to file**: Redirect output: `jot notes search ... > results.md`
- **See Automation Recipes** for export scripts

### Q: What SQL functions are available?

**A**:

- **read_markdown()** - Read markdown files
- **md_stats()** - Get word count, line count, etc.
- See [SQL Functions Reference](sql-functions-reference.md)

### Q: Is Windows supported?

**A**:
Yes! **Fully supported** on Windows 10+. See:

- Use forward slashes in paths: `C:/Users/Me/Notes`
- Or escape backslashes: `C:\\Users\\Me\\Notes`
- Platform-specific issues section above

### Q: Where are my notes stored?

**A**:

- **Notes stay in original location** - Jot doesn't move them
- **Configuration**: `~/.config/jot/config.json`
- **All original files**: Untouched in your notebook directory

### Q: Can I run SQL queries on encrypted notes?

**A**:

- **Encrypted files**: Must be decrypted first (done at OS level)
- **Encrypted filesystems**: Supported if mounted and readable
- **Encrypted text**: Jot sees encrypted content as regular text

### Q: How do I troubleshoot a slow query?

**A**:

1. **Check system resources**: `top` or Task Manager
2. **Simplify the query**: Remove JOINs, aggregate functions
3. **Add LIMIT**: See intermediate results faster
4. **Check file size**: Very large notebooks take time
5. **See Performance Issues section** above

---

## Getting Help

If your issue isn't covered here:

1. **Check the documentation**: [Getting Started Guide](getting-started-power-users.md), [SQL Quick Reference](sql-quick-reference.md)
2. **Read the CLI help**: `jot --help`, `jot notes search --help`
3. **Try the examples**: [Automation Recipes](automation-recipes.md) has working examples
4. **Report bugs**: GitHub issues with command and error output

---

## Success Checklist

You've solved your problem when:

- ✅ Command runs without errors
- ✅ Results match your expectations
- ✅ Performance is acceptable for your use case
- ✅ You understand the root cause of the issue
- ✅ You've documented the solution for next time (if applicable)

Next steps:

- Explore [Advanced Workflows](automation-recipes.md) for automation
- Master [SQL Queries](sql-quick-reference.md) for custom analysis
- Build a [Personal Knowledge Base](getting-started-power-users.md) with Jot
