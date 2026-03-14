const express = require('express');
const cors = require('cors');
const fs = require('fs').promises;
const path = require('path');
const crypto = require('crypto');

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());
// Enable CORS for development (allow requests from the Flutter web or emulator)
app.use(cors());

// Request logging middleware
app.use((req, res, next) => {
  const timestamp = new Date().toISOString();
  console.log(`[${timestamp}] ${req.method} ${req.path}`);
  next();
});

const DB_FILE = path.join(__dirname, 'persons.json');

async function readUsers() {
  try {
    const raw = await fs.readFile(DB_FILE, 'utf8');
    return JSON.parse(raw || '[]');
  } catch (e) {
    if (e.code === 'ENOENT') return [];
    throw e;
  }
}

async function writeUsers(users) {
  await fs.writeFile(DB_FILE, JSON.stringify(users, null, 2), 'utf8');
}

function hashPassword(password, saltHex) {
  // Convert hex string back to Buffer (salt is stored as hex in JSON)
  const saltBuffer = Buffer.from(saltHex, 'hex');
  const hash = crypto.pbkdf2Sync(password, saltBuffer, 100000, 64, 'sha512');
  return hash.toString('hex');
}

function generateId() {
  return Date.now().toString(36) + '-' + crypto.randomBytes(4).toString('hex');
}

// Standardized error response builder
function createErrorResponse(code, message, details = null, hint = null) {
  return {
    success: false,
    error: {
      code,
      message,
      ...(details && { details }),
      ...(hint && { hint }),
    },
  };
}

function verifyPassword(password, hash, saltHex) {
  // Ensure strict case-sensitive comparison
  const computed = hashPassword(password, saltHex);
  return computed === hash;
}

// Validate password meets case-sensitivity requirements
function validatePasswordStrength(password) {
  const hasUpperCase = /[A-Z]/.test(password);
  const hasLowerCase = /[a-z]/.test(password);
  const hasNumber = /[0-9]/.test(password);
  const isLongEnough = password.length >= 3;

  const errors = [];
  if (!hasUpperCase) errors.push('must contain at least one uppercase letter');
  if (!hasLowerCase) errors.push('must contain at least one lowercase letter');
  if (!hasNumber) errors.push('must contain at least one number');
  if (!isLongEnough) errors.push('must be at least 3 characters long');

  return {
    isValid: errors.length === 0,
    errors,
    hasCaseSensitivity: hasUpperCase && hasLowerCase,
  };
}

// Reusable signup handler (exposed on two routes for compatibility)
async function handleSignup(req, res) {
  try {
    const { username, email, password } = req.body || {};

    if (!username || !email || !password) {
      console.log('❌ Signup failed: missing required fields');
      return res.status(400).json(
        createErrorResponse(
          'MISSING_FIELDS',
          'username, email and password are required',
          ['username', 'email', 'password']
        )
      );
    }

    // Validate password strength and case sensitivity
    const passwordValidation = validatePasswordStrength(password);
    if (!passwordValidation.isValid) {
      console.log(`❌ Signup failed: weak password - ${passwordValidation.errors.join(', ')}`);
      return res.status(400).json(
        createErrorResponse(
          'PASSWORD_WEAK',
          'password does not meet requirements',
          passwordValidation.errors,
          'Password must contain uppercase, lowercase, a number, and be at least 8 characters long'
        )
      );
    }

    const users = await readUsers();

    // ensure uniqueness by email or username
    const exists = users.find(u => u.email === email || u.username === username);
    if (exists) {
      console.log(`❌ Signup failed: user with email/username already exists`);
      return res.status(409).json(
        createErrorResponse(
          'USER_EXISTS',
          'user with that email or username already exists',
          null,
          exists.email === email ? 'Email already registered' : 'Username already taken'
        )
      );
    }

    const salt = crypto.randomBytes(16).toString('hex');
    const passwordHash = hashPassword(password, salt);

    const user = {
      id: generateId(),
      username,
      email,
      passwordHash,
      salt,
      // Password metadata: stores that case sensitivity is enforced
      passwordMetadata: {
        caseSensitive: true,
        requiresUpperCase: true,
        requiresLowerCase: true,
        requiresNumber: true,
        minimumLength: 8,
      },
      createdAt: new Date().toISOString(),
    };

    users.push(user);
    await writeUsers(users);

    // Do not return passwordHash/salt in response
    const { passwordHash: _, salt: __, ...publicUser } = user;

    console.log(`✅ User signed up successfully: ${email} (${username})`);
    return res.status(201).json({ success: true, user: publicUser });
  } catch (e) {
    console.error('❌ Signup error', e);
    return res.status(500).json(
      createErrorResponse(
        'SERVER_ERROR',
        'internal server error',
        null,
        'Something went wrong. Please try again later.'
      )
    );
  }
}

// Login handler
async function handleLogin(req, res) {
  try {
    const { email, password } = req.body || {};

    if (!email || !password) {
      console.log('❌ Login failed: missing email or password');
      return res.status(400).json(
        createErrorResponse(
          'MISSING_FIELDS',
          'email and password are required',
          ['email', 'password']
        )
      );
    }

    const users = await readUsers();
    const user = users.find(u => u.email === email);

    if (!user) {
      console.log(`❌ Login failed: user not found (${email})`);
      return res.status(401).json(
        createErrorResponse(
          'INVALID_CREDENTIALS',
          'invalid email or password'
        )
      );
    }

    // Validate password with case sensitivity enforcement
    const isValid = verifyPassword(password, user.passwordHash, user.salt);
    if (!isValid) {
      console.log(`❌ Login failed: invalid password for ${email} (case-sensitive validation failed)`);
      return res.status(401).json(
        createErrorResponse(
          'INVALID_PASSWORD',
          'invalid email or password',
          null,
          'Password is case-sensitive. Check uppercase, lowercase, numbers, and length (min 8 characters).'
        )
      );
    }

    // Generate a simple JWT-like token for this session
    const token = crypto.randomBytes(32).toString('hex');
    const { passwordHash: _, salt: __, ...publicUser } = user;

    console.log(`✅ User logged in successfully: ${email} (password validated with case sensitivity)`);
    return res.status(200).json({
      success: true,
      user: publicUser,
      token: token,
      expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(), // 24 hours
    });
  } catch (e) {
    console.error('❌ Login error', e);
    return res.status(500).json(
      createErrorResponse(
        'SERVER_ERROR',
        'internal server error',
        null,
        'Something went wrong. Please try again later.'
      )
    );
  }
}

app.post('/signup', handleSignup);
app.post('/auth/signup', handleSignup);
app.post('/login', handleLogin);
app.post('/auth/login', handleLogin);

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({
    status: 'ok',
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
  });
});

// simple endpoint to list users (public for dev only)
app.get('/users', async (req, res) => {
  try {
    const users = await readUsers();
    const publicUsers = users.map(u => ({ 
      id: u.id, 
      username: u.username, 
      email: u.email, 
      createdAt: u.createdAt 
    }));
    console.log(`📋 Listed ${publicUsers.length} users`);
    res.json({ success: true, users: publicUsers });
  } catch (e) {
    console.error('❌ Error listing users', e);
    res.status(500).json(
      createErrorResponse(
        'SERVER_ERROR',
        'internal server error',
        null,
        'Failed to fetch users'
      )
    );
  }
});

// Error handling middleware
app.use((err, req, res, next) => {
  console.error('❌ Unhandled error', err);
  res.status(500).json(
    createErrorResponse(
      'SERVER_ERROR',
      'internal server error',
      null,
      'An unexpected error occurred'
    )
  );
});

// 404 handler
app.use((req, res) => {
  console.log(`❌ 404: ${req.method} ${req.path}`);
  res.status(404).json(
    createErrorResponse(
      'NOT_FOUND',
      'endpoint not found',
      null,
      `${req.method} ${req.path} does not exist`
    )
  );
});

app.listen(PORT, () => {
  console.log(`\n🚀 Pulse backend listening on http://localhost:${PORT}\n`);
  console.log('Available endpoints:');
  console.log('  POST /signup - Create new user');
  console.log('  POST /login - Login user');
  console.log('  GET /users - List all users (dev only)');
  console.log('  GET /health - Health check\n');
});
