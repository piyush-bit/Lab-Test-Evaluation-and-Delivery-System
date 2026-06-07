/* ----------------------------------------------------
   Teachyst-Inspired Javascript for EUC2/TDES
   Features: Theme toggling (light/dark), Terminal Simulator,
   Lifecycle tabs, CLI command tabs, Schema Explorer.
---------------------------------------------------- */

// ==========================================
// 0. Theme Switcher Logic (Teachyst Style)
// ==========================================
function initTheme() {
    const themeToggleBtn = document.getElementById('theme-toggle');
    const htmlElement = document.documentElement;

    // Check saved theme or system preference
    const savedTheme = localStorage.getItem('theme');
    const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

    if (savedTheme === 'dark' || (!savedTheme && systemPrefersDark)) {
        htmlElement.classList.add('dark');
    } else {
        htmlElement.classList.remove('dark');
    }

    if (themeToggleBtn) {
        themeToggleBtn.addEventListener('click', () => {
            if (htmlElement.classList.contains('dark')) {
                htmlElement.classList.remove('dark');
                localStorage.setItem('theme', 'light');
            } else {
                htmlElement.classList.add('dark');
                localStorage.setItem('theme', 'dark');
            }
        });
    }
}

// ==========================================
// 1. Terminal Command Simulator
// ==========================================
const terminalData = {
    run: `
<div class="term-line"><span class="term-prompt">~ $</span> <span class="typed-text">euc2 run</span></div>
<div class="term-output">
    <span class="status-info">ℹ️ Running public tests in container...</span><br>
    <span class="status-success">✓ TestStackPush (0.01s)</span><br>
    <span class="status-success">✓ TestStackPop (0.01s)</span><br>
    <span class="status-success">✓ TestStackPeek (0.01s)</span><br>
    <span class="status-success">✓ TestStackIsEmpty (0.01s)</span><br>
    <span class="status-success">✓ TestStackSize (0.01s)</span><br>
    <br>
    <span class="final-success">PASS: 5/5 public tests completed successfully.</span>
</div>
    `,
    evaluate: `
<div class="term-line"><span class="term-prompt">~ $</span> <span class="typed-text">euc2 evaluate ./submission.tar.gz</span></div>
<div class="term-output">
    <span class="status-info">ℹ️ Resolving private test packages...</span><br>
    <span class="status-info">ℹ️ Spin up grading runtime using image "lab-go-runner:v1.0"...</span><br>
    <br>
    <span class="status-success">✓ Target: "make test-submission-lifo" (Points: 2/2)</span><br>
    <span class="status-success">✓ Target: "make test-submission-interleaved" (Points: 2/2)</span><br>
    <span class="status-success">✓ Target: "make test-submission-size" (Points: 2/2)</span><br>
    <span class="status-success">✓ Target: "make test-submission-pop-empty" (Points: 2/2)</span><br>
    <span class="status-success">✓ Target: "make test-submission-peek" (Points: 2/2)</span><br>
    <br>
    <pre style="background: rgba(255,255,255,0.03); border: 1px solid rgba(255,255,255,0.08); padding: 10px; border-radius: 6px; color:#818cf8; font-size: 0.75rem;">
{
  "lab_id": "go101-lab01",
  "score": 10,
  "max_score": 10,
  "passed": true,
  "evaluated_at": "2026-06-07T13:22:15Z"
}</pre>
    <br>
    <span class="final-success">EVALUATION COMPLETE: Score: 10/10</span>
</div>
    `
};

function simulateCommand(cmdType) {
    document.getElementById('btn-run-term').classList.remove('active');
    document.getElementById('btn-eval-term').classList.remove('active');

    if (cmdType === 'run') {
        document.getElementById('btn-run-term').classList.add('active');
    } else {
        document.getElementById('btn-eval-term').classList.add('active');
    }

    const termBody = document.getElementById('terminal-content');
    termBody.innerHTML = `<div class="term-line"><span class="term-prompt">~ $</span> <span class="typed-text">typing...</span></div>`;

    setTimeout(() => {
        termBody.innerHTML = terminalData[cmdType];
    }, 450);
}



// ==========================================
// 3. CLI Command Reference Switcher
// ==========================================
const commandData = {
    init: {
        title: "euc2 init",
        desc: "Initialize an exercise by lab ID and version from local packages into a working directory.",
        args: "<code>&lt;id-with-version&gt; &lt;working-directory&gt;</code>",
        example: `
$ euc2 init go101-lab01@1.1.0 ./my-stack-exercise
Exercise initialized in ./my-stack-exercise
$ ls -A ./my-stack-exercise
Makefile          README.md         manifest.json     run               stack.go
        `
    },
    fetch: {
        title: "euc2 fetch",
        desc: "Fetch an exercise from a drive path or remote server registry, unpacking it locally.",
        args: "<code>&lt;id-with-version&gt; [--drive /path] [--remote URL] [--org-id ID]</code>",
        example: `
# Fetch using local drive strategy
$ euc2 fetch go101-lab01 --drive /Volumes/LabShare
Fetching go101-lab01 from drive source...
Saved public archive to cache.

# Fetch from remote server endpoint
$ euc2 fetch go101-lab01 --remote https://tdes.edu --org-id cse-101
Querying server registry for go101-lab01...
Fetched package successfully.
        `
    },
    package: {
        title: "euc2 package",
        desc: "Packages the exercise workspace, building separate public and private archives and caching them.",
        args: "<code>&lt;exercise-directory&gt;</code>",
        example: `
$ euc2 package ./go101-lab01-stack
Packaging exercise: /Users/piyush/Developer/Test-delivery-and-evalution-system/demo_exercises/go101-lab01-stack
Public packages: [/var/tmp/euc2/cache/go101-lab01_1.1.0_public.tar.gz]
Private packages: [/var/tmp/euc2/private_packages/go101-lab01_1.1.0_private.tar.gz]
Exercise packaged and cached successfully.
        `
    },
    run: {
        title: "euc2 run",
        desc: "Executes the public test targets defined for this exercise in manifest.json to check current logic.",
        args: "<code>[exercise-directory] [--local]</code>",
        example: `
# Runs public test targets using Docker image runner
$ euc2 run
ℹ️ Running public tests in container...
✓ TestStackPush (0.01s)
✓ TestStackPop (0.01s)
PASS: 5/5 public tests completed successfully.

# Runs public tests directly on the host system without Docker containerization
$ euc2 run --local
go test -v ./...
=== RUN   TestStackPush
--- PASS: TestStackPush (0.00s)
PASS
ok  	TDES/stack	0.005s
        `
    },
    submit: {
        title: "euc2 submit",
        desc: "Encrypts and packages files listed in include_paths to deliver to a destination.",
        args: "<code>[--drive /path] [--remote URL] --org-id ID --student-id ID</code>",
        example: `
# Create an encrypted envelope on a shared lab drive
$ euc2 submit --drive /Volumes/LabSubmissions --org-id cse-101 --student-id student_99
Submission envelope created successfully.
Encrypted file written to: /Volumes/LabSubmissions/submissions/go101-lab01/envelope_student_99.enc
        `
    },
    evaluate: {
        title: "euc2 evaluate",
        desc: "Grades a student's submission tar archive against private tests in a secure sandbox.",
        args: "<code>&lt;submission-tar&gt; [--private-store Path] [--docker-binary Path] [--output OutputFile.json]</code>",
        example: `
$ euc2 evaluate ./student_99_submission.tar.gz --output result.json
{
  "lab_id": "go101-lab01",
  "score": 10,
  "max_score": 10,
  "passed": true,
  "evaluated_at": "2026-06-07T13:22:15Z"
}
        `
    },
    drive: {
        title: "euc2 drive",
        desc: "Prepares local or remote drive directory structures for lab delivery and encrypted submissions.",
        args: "<code>prepare &lt;path&gt; | prepare-submission &lt;path&gt; --recipient-public-key &lt;key&gt;</code>",
        example: `
# Prepare drive for exercise delivery
$ euc2 drive prepare /Volumes/LabShare
Drive prepared at: /Volumes/LabShare

# Prepare submissions directory with X25519 keys
$ euc2 drive prepare-submission /Volumes/LabShare --recipient-public-key MCowBQYDK2VuAyEA...
Drive submission module prepared at: /Volumes/LabShare
        `
    }
};

function showCommand(cmdName) {
    const tabs = document.querySelectorAll('.cmd-tab');
    tabs.forEach(tab => {
        if (tab.innerText === cmdName) {
            tab.classList.add('active');
        } else {
            tab.classList.remove('active');
        }
    });

    const display = document.getElementById('command-details');
    const data = commandData[cmdName];

    display.innerHTML = `
        <h3>${data.title}</h3>
        <p class="cmd-desc">${data.desc}</p>
        <div class="cmd-args-section">
            <strong>Arguments / Flags:</strong> ${data.args}
        </div>
        <div class="code-block-header">Usage & Command Output</div>
        <pre><code>${data.example.trim()}</code></pre>
    `;
}

// ==========================================
// 4. Interactive Manifest Schema Explorer
// ==========================================
const schemaDoc = {
    lab_id: {
        title: "lab_id",
        desc: "A unique identifier string for the exercise lab. This is machine-readable, lowercase, and matches the directory name prefix (e.g. <code>go101-lab01</code>). Used to pair grading submissions with reference materials."
    },
    title: {
        title: "title",
        desc: "A human-readable label or name for the exercise (e.g. <code>Slice-Backed Stack in Go</code>) shown to students and displayed on course portals."
    },
    version: {
        title: "version",
        desc: "A Semantic Versioning (SemVer) string representing the current release of the lab (e.g. <code>1.1.0</code>). Version shifts indicate adjustments in test scripts, guidelines, or source code boilerplate."
    },
    language: {
        title: "language",
        desc: "Specifies the programming language target of the exercise (e.g., <code>go</code>, <code>c</code>, <code>python</code>, <code>cpp</code>). Helps orchestrate syntax highlights or language-specific compiler actions."
    },
    runner_image: {
        title: "runner_image",
        desc: "The container image name used by the Docker run service during public tests and grading (e.g. <code>lab-go-runner:v1.0</code>). The image must support basic utilities like <code>/bin/sh</code> and <code>make</code>."
    },
    local_entrypoint: {
        title: "local_entrypoint",
        desc: "The script or command executed locally when the student triggers test checking in their IDE (defaults to <code>make test-public</code>)."
    },
    grading: {
        title: "grading",
        desc: "An ordered array of grading target commands. The evaluation system runs each target in sequence, checking for a zero exit status. Points are awarded progressively based on individual target pass status."
    },
    submission: {
        title: "submission",
        desc: "Configures files targeted for submissions and strips hidden test assets. <code>include_paths</code> marks editable student files. <code>private_globs</code> filters security-sensitive items (like reference solutions and private tests) which are stripped from student packages."
    },
    limits: {
        title: "limits",
        desc: "Applies resource boundary conditions to Docker test containers. Enforces maximum memory usage (<code>memory_mb</code>), maximum runtimes per test command (<code>timeout_seconds</code>), and limits process counts (<code>pids_limit</code>)."
    }
};

function initSchemaExplorer() {
    const keys = document.querySelectorAll('.schema-key');
    const descPanel = document.getElementById('schema-desc-panel');

    keys.forEach(keyElement => {
        const target = keyElement.getAttribute('data-target');

        keyElement.addEventListener('mouseover', () => {
            keys.forEach(k => k.classList.remove('selected'));
            keyElement.classList.add('selected');
            updateDescPanel(target);
        });

        keyElement.addEventListener('click', (e) => {
            e.stopPropagation();
            keys.forEach(k => k.classList.remove('selected'));
            keyElement.classList.add('selected');
            updateDescPanel(target);
        });
    });

    function updateDescPanel(keyName) {
        const info = schemaDoc[keyName];
        if (info) {
            descPanel.innerHTML = `
                <h3>💡 ${info.title}</h3>
                <p>${info.desc}</p>
            `;
        }
    }
}

// ==========================================
// 5. Scroll Spy for Docs Table of Contents
// ==========================================
function initScrollSpy() {
    const tocLinks = document.querySelectorAll('.toc-link');
    const sections = document.querySelectorAll('.docs-section-block');

    if (tocLinks.length === 0 || sections.length === 0) return;

    window.addEventListener('scroll', () => {
        let currentSectionId = 'intro';
        
        sections.forEach(section => {
            const sectionTop = section.offsetTop;
            if (window.scrollY >= sectionTop - 160) {
                currentSectionId = section.getAttribute('id');
            }
        });

        tocLinks.forEach(link => {
            link.classList.remove('active');
            if (link.getAttribute('href') === `#${currentSectionId}`) {
                link.classList.add('active');
            }
        });
    });
}

// ==========================================
// Initialization
// ==========================================
document.addEventListener('DOMContentLoaded', () => {
    initTheme();
    simulateCommand('run');
    initSchemaExplorer();
    initScrollSpy();
});
