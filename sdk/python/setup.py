from setuptools import setup, find_packages

setup(
    name="pulse-iit",
    version="0.1.0",
    description="Pulse IIT Python SDK — server-side event tracking com buffering e batch flush",
    long_description=open("README.md", encoding="utf-8").read(),
    long_description_content_type="text/markdown",
    author="Instituto Itinerante de Tecnologia",
    url="https://github.com/apsferreira/pulse-service",
    license="MIT",
    packages=find_packages(exclude=["tests*"]),
    python_requires=">=3.9",
    install_requires=[
        "urllib3>=1.26,<3",
    ],
    extras_require={
        "dev": ["pytest>=7", "responses>=0.25"],
    },
    classifiers=[
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Intended Audience :: Developers",
        "Topic :: Software Development :: Libraries :: Python Modules",
    ],
    project_urls={
        "Source": "https://github.com/apsferreira/pulse-service/tree/main/sdk/python",
        "Tracker": "https://github.com/apsferreira/pulse-service/issues",
    },
)
