import * as React from 'react'

// ADMIN PANEL
export default function Admin() {

    // Название тренировки
    const [title, setTitle] =
        React.useState('')

    // Описание тренировки
    const [description, setDescription] =
        React.useState('')

    // Длительность тренировки
    const [duration, setDuration] =
        React.useState('')

    // Сложность
    const [difficulty, setDifficulty] =
        React.useState('')

    // YouTube ссылка
    const [youtubeLink, setYoutubeLink] =
        React.useState('')

    // Состояние загрузки
    const [loading, setLoading] =
        React.useState(false)

    // Сообщение об ошибке
    const [error, setError] =
        React.useState<string | null>(null)

    // Сообщение об успехе
    const [success, setSuccess] =
        React.useState<string | null>(null)

    // Список тренировок
    const [trainings, setTrainings] =
        React.useState<any[]>([])

    // Список питания
    const [nutrition, setNutrition] =
        React.useState<any[]>([])

    // Поля формы питания
    const [nutritionTitle, setNutritionTitle] =
        React.useState('')

    const [nutritionDescription, setNutritionDescription] =
        React.useState('')

    const [calories, setCalories] =
        React.useState('')

    const [proteins, setProteins] =
        React.useState('')

    const [fats, setFats] =
        React.useState('')

    const [carbs, setCarbs] =
        React.useState('')

    // Загрузка тренировок
    async function loadTrainings() {

        try {

            const response = await fetch(
                '/api/trainings',
            )

            const data = await response.json()

            setTrainings(data)
        }
        catch (err) {

            console.error(err)
        }
    }

    // Загрузка питания
    async function loadNutrition() {

        try {

            const response = await fetch(
                '/api/nutrition',
            )

            const data = await response.json()

            setNutrition(data)

        } catch (err) {

            console.error(err)
        }
    }

    // Удаление тренировки
    async function deleteTraining(
        id: number,
    ) {

        try {

            await fetch(

                `/api/admin/trainings/${id}?telegramId=898030333`,

                {
                    method: 'DELETE',
                },
            )

            // Обновляем список
            loadTrainings()
        }
        catch (err) {

            console.error(err)
        }
    }

    // Удаление питания
    async function deleteNutrition(
        id: number,
    ) {

        try {

            await fetch(

                `/api/admin/nutrition/${id}?telegramId=898030333`,

                {
                    method: 'DELETE',
                },
            )

            loadNutrition()

        } catch (err) {

            console.error(err)
        }
    }

    // Создание тренировки
    async function createTraining() {

        // Сброс сообщений
        setError(null)
        setSuccess(null)

        // Проверка обязательных полей
        if (
            !title ||
            !description ||
            !duration
        ) {

            setError(
                'Заполните все обязательные поля',
            )

            return
        }

        try {

            // Включаем loading
            setLoading(true)

            // Отправляем POST запрос
            const response = await fetch(

                '/api/admin/trainings',

                {
                    method: 'POST',

                    headers: {
                        'Content-Type': 'application/json',
                    },

                    body: JSON.stringify({

                        telegram_id: 898030333,

                        title,

                        description,

                        duration: Number(duration),

                        difficulty,

                        youtube_link: youtubeLink,
                    }),
                },
            )

            // Если ошибка API
            if (!response.ok) {

                throw new Error(
                    'Ошибка при создании тренировки',
                )
            }

            // Успешное сообщение
            setSuccess(
                '✅ Тренировка успешно добавлена',
            )

            // Очистка формы
            setTitle('')
            setDescription('')
            setDuration('')
            setDifficulty('')
            setYoutubeLink('')

            // Обновляем список
            loadTrainings()
        }
        catch (err: any) {

            setError(
                err.message,
            )
        }
        finally {

            // Выключаем loading
            setLoading(false)
        }
    }

    // Создание питания
    async function createNutrition() {

        try {

            const response = await fetch(

                '/api/admin/nutrition',

                {
                    method: 'POST',

                    headers: {
                        'Content-Type': 'application/json',
                    },

                    body: JSON.stringify({

                        telegram_id: 898030333,

                        title: nutritionTitle,

                        description: nutritionDescription,

                        calories: Number(calories),

                        proteins: Number(proteins),

                        fats: Number(fats),

                        carbs: Number(carbs),

                        category_id: 5,
                    }),
                },
            )

            if (!response.ok) {

                throw new Error(
                    'Ошибка создания питания',
                )
            }

            // Очистка
            setNutritionTitle('')
            setNutritionDescription('')
            setCalories('')
            setProteins('')
            setFats('')
            setCarbs('')

            // Обновляем список
            loadNutrition()

        } catch (err) {

            console.error(err)
        }
    }

    // Загружаем тренировки при открытии страницы
    React.useEffect(() => {

        loadTrainings()

        loadNutrition()

    }, [])

    // RENDER
    return (

        <div className="stack">

            <section className="card">

                {/* HEADER */}

                <div className="section-head">

                    <h2>
                        👑 Admin Panel
                    </h2>

                </div>

                <p className="muted">

                    Управление тренировками платформы

                </p>

                {/* SUCCESS */}

                {success && (

                    <div className="success-box">
                        {success}
                    </div>
                )}

                {/* ERROR */}

                {error && (

                    <div className="error-box">
                        {error}
                    </div>
                )}

                {/* FORM */}

                <div className="form-group">

                    {/* TITLE */}

                    <input
                        className="input"
                        placeholder="Название тренировки"
                        value={title}
                        onChange={(e) =>
                            setTitle(
                                e.target.value,
                            )
                        }
                    />

                    {/* DESCRIPTION */}

                    <textarea
                        className="input"
                        placeholder="Описание тренировки"
                        value={description}
                        onChange={(e) =>
                            setDescription(
                                e.target.value,
                            )
                        }
                    />

                    {/* DURATION */}

                    <input
                        className="input"
                        placeholder="Длительность (мин)"
                        value={duration}
                        onChange={(e) =>
                            setDuration(
                                e.target.value,
                            )
                        }
                    />

                    {/* DIFFICULTY */}

                    <input
                        className="input"
                        placeholder="Сложность"
                        value={difficulty}
                        onChange={(e) =>
                            setDifficulty(
                                e.target.value,
                            )
                        }
                    />

                    {/* YOUTUBE */}

                    <input
                        className="input"
                        placeholder="YouTube ссылка"
                        value={youtubeLink}
                        onChange={(e) =>
                            setYoutubeLink(
                                e.target.value,
                            )
                        }
                    />

                    {/* BUTTON */}

                    <button
                        className="primary-btn"
                        onClick={createTraining}
                        disabled={loading}
                    >

                        {loading
                            ? '⏳ Сохранение...'
                            : '➕ Добавить тренировку'}

                    </button>

                </div>

            </section>

            <section className="card">

                <div className="section-head">

                    <h2>
                        🍎 Питание
                    </h2>

                </div>

                <div className="form-group">

                    <input
                        className="input"
                        placeholder="Название"
                        value={nutritionTitle}
                        onChange={(e) =>
                            setNutritionTitle(
                                e.target.value,
                            )
                        }
                    />

                    <textarea
                        className="input"
                        placeholder="Описание"
                        value={nutritionDescription}
                        onChange={(e) =>
                            setNutritionDescription(
                                e.target.value,
                            )
                        }
                    />

                    <input
                        className="input"
                        placeholder="Калории"
                        value={calories}
                        onChange={(e) =>
                            setCalories(
                                e.target.value,
                            )
                        }
                    />

                    <input
                        className="input"
                        placeholder="Белки"
                        value={proteins}
                        onChange={(e) =>
                            setProteins(
                                e.target.value,
                            )
                        }
                    />

                    <input
                        className="input"
                        placeholder="Жиры"
                        value={fats}
                        onChange={(e) =>
                            setFats(
                                e.target.value,
                            )
                        }
                    />

                    <input
                        className="input"
                        placeholder="Углеводы"
                        value={carbs}
                        onChange={(e) =>
                            setCarbs(
                                e.target.value,
                            )
                        }
                    />

                    <button
                        className="primary-btn"
                        onClick={createNutrition}
                    >

                        ➕ Добавить питание

                    </button>

                </div>

            </section>

            <section className="card">

                <div className="section-head">

                    <h2>
                        📋 Список питания
                    </h2>

                </div>

                <div className="training-list">

                    {nutrition.map((n) => (

                        <div
                            key={n.ID}
                            className="training-card"
                        >

                            <h3>
                                {n.Title}
                            </h3>

                            <p>
                                {n.Description}
                            </p>

                            <p>
                                🔥 {n.Calories} ккал
                            </p>

                            <p>
                                Б: {n.Proteins}
                                |
                                Ж: {n.Fats}
                                |
                                У: {n.Carbs}
                            </p>

                            <button
                                className="primary-btn"
                                onClick={() =>
                                    deleteNutrition(
                                        n.ID,
                                    )
                                }
                            >

                                🗑 Удалить

                            </button>

                        </div>
                    ))}

                </div>

            </section>

            {/* СПИСОК ТРЕНИРОВОК */}

            <section className="card">

                <div className="section-head">

                    <h2>
                        📋 Список тренировок
                    </h2>

                </div>

                <div className="training-list">

                    {trainings.map((t) => (

                        <div
                            key={t.ID}
                            className="training-card"
                        >

                            <h3>
                                {t.Title}
                            </h3>

                            <p>
                                {t.Description}
                            </p>

                            <p>
                                ⏱ {t.Duration} мин
                            </p>

                            <button
                                className="primary-btn"
                                onClick={() =>
                                    deleteTraining(
                                        t.ID,
                                    )
                                }
                            >

                                🗑 Удалить

                            </button>

                        </div>
                    ))}

                </div>

            </section>

        </div>
    )
}