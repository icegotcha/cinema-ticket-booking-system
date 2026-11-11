const databaseName = process.env.MONGODB_DATABASE;

if (!databaseName) {
  throw new Error("MONGODB_DATABASE is required");
}

const database = db.getSiblingDB(databaseName);
const movies = database.movies.find({}, { _id: 1 }).sort({ rating: -1 }).limit(20).toArray();

if (movies.length === 0) {
  throw new Error("Cannot seed showtimes because no movies were found");
}

// Keep the development seed useful over time by scheduling the next seven days.
// The values are BSON timestamps because that is the type used by ShowTime.StartTime.
const showHours = [10, 13, 16, 19];
const firstDay = new Date();
firstDay.setUTCHours(0, 0, 0, 0);

const showtimes = [];

for (const movie of movies) {
  for (let day = 0; day < 7; day += 1) {
    for (let slot = 0; slot < showHours.length; slot += 1) {
      const startTime = new Date(firstDay);
      startTime.setUTCDate(firstDay.getUTCDate() + day);
      startTime.setUTCHours(showHours[slot]);

      showtimes.push({
        movie_id: movie._id,
        start_time: Timestamp({
          t: Math.floor(startTime.getTime() / 1000),
          i: slot + 1,
        }),
      });
    }
  }
}

database.showtimes.deleteMany({});
database.showtimes.insertMany(showtimes);

print(`Seeded ${showtimes.length} showtimes for ${movies.length} movies`);
