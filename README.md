# Marketplace Backend
Open-source backend for a global buy/sell marketplace with Nostr and LNBits.

## Setup
1. Install Docker and Go.
2. Clone the repo: `git clone https://github.com/your-org/marketplace-backend`.
3. Set environment variables: `export JWT_SECRET=your-secret-key LNBITS_API_KEY=your-key`.
4. Run: `docker-compose up`.

## Services
- **AuthService**: Nostr-based auth with JWT (`/auth/login`).
- **ListingService**: Manages listings and LNBits invoices (`/listings/:id`).

## Contributing
See [CONTRIBUTING.md](CONTRIBUTING.md).

## License
MIT License
